package orderer

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/binaries"
	"github.com/chainlaunch/chainlaunch/pkg/db"
	fabricservice "github.com/chainlaunch/chainlaunch/pkg/fabric/service"
	kmodels "github.com/chainlaunch/chainlaunch/pkg/keymanagement/models"
	keymanagement "github.com/chainlaunch/chainlaunch/pkg/keymanagement/service"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/chainlaunch/chainlaunch/pkg/nodes/orderer/osnadmin"
	"github.com/chainlaunch/chainlaunch/pkg/nodes/types"
)

// LocalOrderer represents a local Fabric orderer node
type LocalOrderer struct {
	mspID          string
	db             *db.Queries
	opts           StartOrdererOpts
	mode           string
	org            *fabricservice.OrganizationDTO
	organizationID int64
	orgService     *fabricservice.OrganizationService
	keyService     *keymanagement.KeyManagementService
	nodeID         int64
	logger         *logger.Logger
}

// NewLocalOrderer creates a new LocalOrderer instance
func NewLocalOrderer(
	mspID string,
	db *db.Queries,
	opts StartOrdererOpts,
	mode string,
	org *fabricservice.OrganizationDTO,
	organizationID int64,
	orgService *fabricservice.OrganizationService,
	keyService *keymanagement.KeyManagementService,
	nodeID int64,
	logger *logger.Logger,
) *LocalOrderer {
	return &LocalOrderer{
		mspID:          mspID,
		db:             db,
		opts:           opts,
		mode:           mode,
		org:            org,
		organizationID: organizationID,
		orgService:     orgService,
		keyService:     keyService,
		nodeID:         nodeID,
		logger:         logger,
	}
}

// getServiceName returns the systemd service name
func (o *LocalOrderer) getServiceName() string {
	return fmt.Sprintf("fabric-orderer-%s", strings.ReplaceAll(strings.ToLower(o.opts.ID), " ", "-"))
}

// getLaunchdServiceName returns the launchd service name
func (o *LocalOrderer) getLaunchdServiceName() string {
	return fmt.Sprintf("ai.chainlaunch.orderer.%s.%s",
		strings.ToLower(o.org.MspID),
		strings.ReplaceAll(strings.ToLower(o.opts.ID), " ", "-"))
}

// getServiceFilePath returns the systemd service file path
func (o *LocalOrderer) getServiceFilePath() string {
	return fmt.Sprintf("/etc/systemd/system/%s.service", o.getServiceName())
}

// getLaunchdPlistPath returns the launchd plist file path
func (o *LocalOrderer) getLaunchdPlistPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "Library/LaunchAgents", o.getLaunchdServiceName()+".plist")
}

// GetStdOutPath returns the path to the stdout log file
func (o *LocalOrderer) GetStdOutPath() string {
	homeDir, _ := os.UserHomeDir()
	dirPath := filepath.Join(homeDir, ".chainlaunch/orderers",
		strings.ReplaceAll(strings.ToLower(o.opts.ID), " ", "-"))
	return filepath.Join(dirPath, o.getServiceName()+".log")
}

// findOrdererBinary finds the orderer binary in PATH
func (o *LocalOrderer) findOrdererBinary() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	downloader, err := binaries.NewBinaryDownloader(homeDir)
	if err != nil {
		return "", fmt.Errorf("failed to create binary downloader: %w", err)
	}

	return downloader.GetBinaryPath(binaries.OrdererBinary, o.opts.Version)
}

// Start starts the orderer node
func (o *LocalOrderer) Start() (interface{}, error) {
	o.logger.Info("Starting orderer", "opts", o.opts)
	slugifiedID := strings.ReplaceAll(strings.ToLower(o.opts.ID), " ", "-")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	dirPath := filepath.Join(homeDir, ".chainlaunch/orderers", slugifiedID)
	mspConfigPath := filepath.Join(dirPath, "config")
	dataConfigPath := filepath.Join(dirPath, "data")

	// Find orderer binary
	ordererBinary, err := o.findOrdererBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to find orderer binary: %w", err)
	}

	// Build command and environment
	cmd := fmt.Sprintf("%s", ordererBinary)
	env := o.buildOrdererEnvironment(mspConfigPath)

	o.logger.Debug("Starting orderer",
		"mode", o.mode,
		"cmd", cmd,
		"env", env,
		"dirPath", dirPath,
	)

	switch o.mode {
	case "service":
		return o.startService(cmd, env, dirPath)
	case "docker":
		return o.startDocker(env, mspConfigPath, dataConfigPath)
	default:
		return nil, fmt.Errorf("invalid mode: %s", o.mode)
	}
}

// Stop stops the orderer node
func (o *LocalOrderer) Stop() error {
	o.logger.Info("Stopping orderer", "opts", o.opts)

	switch o.mode {
	case "service":
		platform := runtime.GOOS
		switch platform {
		case "linux":
			return o.stopSystemdService()
		case "darwin":
			return o.stopLaunchdService()
		default:
			return fmt.Errorf("unsupported platform for service mode: %s", platform)
		}
	case "docker":
		return o.stopDocker()
	default:
		return fmt.Errorf("invalid mode: %s", o.mode)
	}
}

// buildOrdererEnvironment builds the environment variables for the orderer
func (o *LocalOrderer) buildOrdererEnvironment(mspConfigPath string) map[string]string {
	env := make(map[string]string)

	// Add custom environment variables from opts
	for k, v := range o.opts.Env {
		env[k] = v
	}

	// Add required environment variables
	env["FABRIC_CFG_PATH"] = mspConfigPath
	env["ORDERER_ADMIN_TLS_CLIENTROOTCAS"] = filepath.Join(mspConfigPath, "tlscacerts/cacert.pem")
	env["ORDERER_ADMIN_TLS_PRIVATEKEY"] = filepath.Join(mspConfigPath, "tls.key")
	env["ORDERER_ADMIN_TLS_CERTIFICATE"] = filepath.Join(mspConfigPath, "tls.crt")
	env["ORDERER_ADMIN_TLS_ROOTCAS"] = filepath.Join(mspConfigPath, "tlscacerts/cacert.pem")
	env["ORDERER_FILELEDGER_LOCATION"] = filepath.Join(mspConfigPath, "data")
	env["ORDERER_GENERAL_CLUSTER_CLIENTCERTIFICATE"] = filepath.Join(mspConfigPath, "tls.crt")
	env["ORDERER_GENERAL_CLUSTER_CLIENTPRIVATEKEY"] = filepath.Join(mspConfigPath, "tls.key")
	env["ORDERER_GENERAL_CLUSTER_ROOTCAS"] = filepath.Join(mspConfigPath, "tlscacerts/cacert.pem")
	env["ORDERER_GENERAL_LOCALMSPDIR"] = mspConfigPath
	env["ORDERER_GENERAL_TLS_CLIENTROOTCAS"] = filepath.Join(mspConfigPath, "tlscacerts/cacert.pem")
	env["ORDERER_GENERAL_TLS_CERTIFICATE"] = filepath.Join(mspConfigPath, "tls.crt")
	env["ORDERER_GENERAL_TLS_PRIVATEKEY"] = filepath.Join(mspConfigPath, "tls.key")
	env["ORDERER_GENERAL_TLS_ROOTCAS"] = filepath.Join(mspConfigPath, "tlscacerts/cacert.pem")
	env["ORDERER_ADMIN_LISTENADDRESS"] = o.opts.AdminAddress
	env["ORDERER_GENERAL_LISTENADDRESS"] = strings.Split(o.opts.ListenAddress, ":")[0]
	env["ORDERER_OPERATIONS_LISTENADDRESS"] = o.opts.OperationsListenAddress
	env["ORDERER_GENERAL_LOCALMSPID"] = o.mspID
	env["ORDERER_GENERAL_LISTENPORT"] = strings.Split(o.opts.ListenAddress, ":")[1]
	env["ORDERER_ADMIN_TLS_ENABLED"] = "true"
	env["ORDERER_CHANNELPARTICIPATION_ENABLED"] = "true"
	env["ORDERER_GENERAL_BOOTSTRAPMETHOD"] = "none"
	env["ORDERER_GENERAL_GENESISPROFILE"] = "initial"
	env["ORDERER_GENERAL_LEDGERTYPE"] = "file"
	env["FABRIC_LOGGING_SPEC"] = "info"
	env["ORDERER_GENERAL_TLS_CLIENTAUTHREQUIRED"] = "false"
	env["ORDERER_GENERAL_TLS_ENABLED"] = "true"
	env["ORDERER_METRICS_PROVIDER"] = "prometheus"
	env["ORDERER_OPERATIONS_TLS_ENABLED"] = "false"

	return env
}

// TailLogs tails the logs of the orderer service
func (o *LocalOrderer) TailLogs(ctx context.Context, tail int, follow bool) (<-chan string, error) {
	logChan := make(chan string, 100)
	logPath := o.GetStdOutPath()

	// Check if log file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		close(logChan)
		return logChan, fmt.Errorf("log file does not exist: %s", logPath)
	}

	// Start goroutine to tail logs
	go func() {
		defer close(logChan)

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			// For Windows, use PowerShell Get-Content
			if follow {
				cmd = exec.Command("powershell", "Get-Content", "-Path", logPath, "-Tail", fmt.Sprintf("%d", tail), "-Wait")
			} else {
				cmd = exec.Command("powershell", "Get-Content", "-Path", logPath, "-Tail", fmt.Sprintf("%d", tail))
			}
		} else {
			// For Unix-like systems, use tail command
			if follow {
				cmd = exec.Command("tail", "-n", fmt.Sprintf("%d", tail), "-f", logPath)
			} else {
				cmd = exec.Command("tail", "-n", fmt.Sprintf("%d", tail), logPath)
			}
		}

		// Create pipe for reading command output
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			o.logger.Error("Failed to create stdout pipe", "error", err)
			return
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			o.logger.Error("Failed to start tail command", "error", err)
			return
		}

		// Create scanner to read output line by line
		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanLines)

		// Read lines and send to channel
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				// Context cancelled, stop tailing
				cmd.Process.Kill()
				return
			case logChan <- scanner.Text():
				// Line sent successfully
			}
		}

		// Wait for command to complete
		if err := cmd.Wait(); err != nil {
			if ctx.Err() == nil { // Only log error if context wasn't cancelled
				o.logger.Error("Tail command failed", "error", err)
			}
		}
	}()

	return logChan, nil
}

// Init initializes the orderer configuration
func (o *LocalOrderer) Init() (interface{}, error) {
	ctx := context.Background()
	// Get node from database
	node, err := o.db.GetNode(ctx, o.nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	o.logger.Info("Initializing orderer",
		"opts", o.opts,
		"node", node,
		"orgID", o.organizationID,
		"nodeID", o.nodeID,
	)

	// Get organization
	org, err := o.orgService.GetOrganization(ctx, o.organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	signCAKeyDB, err := o.keyService.GetKey(ctx, int(org.SignKeyID.Int64))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve sign CA cert: %w", err)
	}

	tlsCAKeyDB, err := o.keyService.GetKey(ctx, int(org.TlsRootKeyID.Int64))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve TLS CA cert: %w", err)
	}

	isCA := 0
	description := "Sign key for " + o.opts.ID
	curveP256 := kmodels.ECCurveP256
	providerID := 1

	// Create Sign Key
	signKeyDB, err := o.keyService.CreateKey(ctx, kmodels.CreateKeyRequest{
		Algorithm:   kmodels.KeyAlgorithmEC,
		Name:        o.opts.ID,
		IsCA:        &isCA,
		Description: &description,
		Curve:       &curveP256,
		ProviderID:  &providerID,
	}, int(org.SignKeyID.Int64))
	if err != nil {
		return nil, fmt.Errorf("failed to create sign key: %w", err)
	}

	// Sign Sign Key
	signKeyDB, err = o.keyService.SignCertificate(ctx, signKeyDB.ID, signCAKeyDB.ID, kmodels.CertificateRequest{
		CommonName:         o.opts.ID,
		Organization:       []string{org.MspID},
		OrganizationalUnit: []string{"orderer"},
		DNSNames:           []string{o.opts.ID},
		IsCA:               true,
		KeyUsage:           x509.KeyUsageCertSign,
		ExtKeyUsage:        []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign sign key: %w", err)
	}

	signKey, err := o.keyService.GetDecryptedPrivateKey(int(signKeyDB.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to get sign private key: %w", err)
	}

	// Create TLS key
	tlsKeyDB, err := o.keyService.CreateKey(ctx, kmodels.CreateKeyRequest{
		Algorithm:   kmodels.KeyAlgorithmEC,
		Name:        o.opts.ID,
		IsCA:        &isCA,
		Description: &description,
		Curve:       &curveP256,
		ProviderID:  &providerID,
	}, int(org.SignKeyID.Int64))
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS key: %w", err)
	}

	// Sign TLS certificates
	validFor := kmodels.Duration(time.Hour * 24 * 365)
	// Add localhost and 127.0.0.1 to domain names if not present
	domainNames := o.opts.DomainNames
	hasLocalhost := false
	hasLoopback := false
	var ipAddresses []net.IP
	var domains []string
	for _, domain := range domainNames {
		if domain == "localhost" {
			hasLocalhost = true
			domains = append(domains, domain)
			continue
		}
		if domain == "127.0.0.1" {
			hasLoopback = true
			ipAddresses = append(ipAddresses, net.ParseIP(domain))
			continue
		}
		if ip := net.ParseIP(domain); ip != nil {
			ipAddresses = append(ipAddresses, ip)
		} else {
			domains = append(domains, domain)
		}
	}
	if !hasLocalhost {
		domains = append(domains, "localhost")
	}
	if !hasLoopback {
		ipAddresses = append(ipAddresses, net.ParseIP("127.0.0.1"))
	}
	o.opts.DomainNames = domains

	tlsKeyDB, err = o.keyService.SignCertificate(ctx, tlsKeyDB.ID, tlsCAKeyDB.ID, kmodels.CertificateRequest{
		CommonName:         o.opts.ID,
		Organization:       []string{org.MspID},
		OrganizationalUnit: []string{"orderer"},
		DNSNames:           domains,
		IPAddresses:        ipAddresses,
		IsCA:               true,
		ValidFor:           validFor,
		KeyUsage:           x509.KeyUsageCertSign,
		ExtKeyUsage:        []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign TLS certificate: %w", err)
	}

	tlsKey, err := o.keyService.GetDecryptedPrivateKey(int(tlsKeyDB.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to get TLS private key: %w", err)
	}

	// Create directory structure
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	slugifiedID := strings.ReplaceAll(strings.ToLower(o.opts.ID), " ", "-")
	dirPath := filepath.Join(homeDir, ".chainlaunch", "orderers", slugifiedID)
	dataConfigPath := filepath.Join(dirPath, "data")
	mspConfigPath := filepath.Join(dirPath, "config")

	// Create directories
	if err := os.MkdirAll(dataConfigPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	if err := os.MkdirAll(mspConfigPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create msp directory: %w", err)
	}

	// Write certificates and keys
	if err := o.writeCertificatesAndKeys(mspConfigPath, tlsKeyDB, signKeyDB, tlsKey, signKey, signCAKeyDB, tlsCAKeyDB); err != nil {
		return nil, fmt.Errorf("failed to write certificates and keys: %w", err)
	}

	// Write config files
	if err := o.writeConfigFiles(mspConfigPath, dataConfigPath); err != nil {
		return nil, fmt.Errorf("failed to write config files: %w", err)
	}

	return &types.FabricOrdererDeploymentConfig{
		BaseDeploymentConfig: types.BaseDeploymentConfig{
			Type: "fabric-orderer",
			Mode: o.mode,
		},
		OrganizationID:          o.organizationID,
		MSPID:                   o.mspID,
		ListenAddress:           o.opts.ListenAddress,
		AdminAddress:            o.opts.AdminAddress,
		OperationsListenAddress: o.opts.OperationsListenAddress,
		ExternalEndpoint:        o.opts.ExternalEndpoint,
		DomainNames:             o.opts.DomainNames,
		SignCert:                *signKeyDB.Certificate,
		TLSCert:                 *tlsKeyDB.Certificate,
		CACert:                  *signCAKeyDB.Certificate,
		TLSCACert:               *tlsCAKeyDB.Certificate,
		SignKeyID:               int64(signKeyDB.ID),
		TLSKeyID:                int64(tlsKeyDB.ID),
	}, nil
}

// writeCertificatesAndKeys writes the certificates and keys to the MSP directory structure
func (o *LocalOrderer) writeCertificatesAndKeys(
	mspConfigPath string,
	tlsCert *kmodels.KeyResponse,
	signCert *kmodels.KeyResponse,
	tlsKey string,
	signKey string,
	signCACert *kmodels.KeyResponse,
	tlsCACert *kmodels.KeyResponse,
) error {
	// Write TLS certificates and keys
	if err := os.WriteFile(filepath.Join(mspConfigPath, "tls.crt"), []byte(*tlsCert.Certificate), 0644); err != nil {
		return fmt.Errorf("failed to write TLS certificate: %w", err)
	}
	if err := os.WriteFile(filepath.Join(mspConfigPath, "tls.key"), []byte(tlsKey), 0600); err != nil {
		return fmt.Errorf("failed to write TLS key: %w", err)
	}

	// Create and write to signcerts directory
	signcertsPath := filepath.Join(mspConfigPath, "signcerts")
	if err := os.MkdirAll(signcertsPath, 0755); err != nil {
		return fmt.Errorf("failed to create signcerts directory: %w", err)
	}
	if err := os.WriteFile(filepath.Join(signcertsPath, "cert.pem"), []byte(*signCert.Certificate), 0644); err != nil {
		return fmt.Errorf("failed to write signing certificate: %w", err)
	}

	// Write root CA certificate
	if err := os.WriteFile(filepath.Join(mspConfigPath, "cacert.pem"), []byte(*signCACert.Certificate), 0644); err != nil {
		return fmt.Errorf("failed to write CA certificate: %w", err)
	}

	// Create and write to cacerts directory
	cacertsPath := filepath.Join(mspConfigPath, "cacerts")
	if err := os.MkdirAll(cacertsPath, 0755); err != nil {
		return fmt.Errorf("failed to create cacerts directory: %w", err)
	}
	if err := os.WriteFile(filepath.Join(cacertsPath, "cacert.pem"), []byte(*signCACert.Certificate), 0644); err != nil {
		return fmt.Errorf("failed to write CA certificate to cacerts: %w", err)
	}

	// Create and write to tlscacerts directory
	tlscacertsPath := filepath.Join(mspConfigPath, "tlscacerts")
	if err := os.MkdirAll(tlscacertsPath, 0755); err != nil {
		return fmt.Errorf("failed to create tlscacerts directory: %w", err)
	}
	if err := os.WriteFile(filepath.Join(tlscacertsPath, "cacert.pem"), []byte(*tlsCACert.Certificate), 0644); err != nil {
		return fmt.Errorf("failed to write TLS CA certificate: %w", err)
	}

	// Create and write to keystore directory
	keystorePath := filepath.Join(mspConfigPath, "keystore")
	if err := os.MkdirAll(keystorePath, 0755); err != nil {
		return fmt.Errorf("failed to create keystore directory: %w", err)
	}
	if err := os.WriteFile(filepath.Join(keystorePath, "key.pem"), []byte(signKey), 0600); err != nil {
		return fmt.Errorf("failed to write signing key: %w", err)
	}

	return nil
}

// writeConfigFiles writes the config.yaml and orderer.yaml files
func (o *LocalOrderer) writeConfigFiles(mspConfigPath, dataConfigPath string) error {
	// Write config.yaml
	configYamlContent := `NodeOUs:
  Enable: true
  ClientOUIdentifier:
    Certificate: cacerts/cacert.pem
    OrganizationalUnitIdentifier: client
  PeerOUIdentifier:
    Certificate: cacerts/cacert.pem
    OrganizationalUnitIdentifier: peer
  AdminOUIdentifier:
    Certificate: cacerts/cacert.pem
    OrganizationalUnitIdentifier: admin
  OrdererOUIdentifier:
    Certificate: cacerts/cacert.pem
    OrganizationalUnitIdentifier: orderer
`
	if err := os.WriteFile(filepath.Join(mspConfigPath, "config.yaml"), []byte(configYamlContent), 0644); err != nil {
		return fmt.Errorf("failed to write config.yaml: %w", err)
	}

	// Write orderer.yaml
	ordererYamlTemplate := `
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

---
################################################################################
#
#   Orderer Configuration
#
#   - This controls the type and configuration of the orderer.
#
################################################################################
General:
    # Listen address: The IP on which to bind to listen.
    ListenAddress: 127.0.0.1

    # Listen port: The port on which to bind to listen.
    ListenPort: 7050

    # TLS: TLS settings for the GRPC server.
    TLS:
        # Require server-side TLS
        Enabled: false
        # PrivateKey governs the file location of the private key of the TLS certificate.
        PrivateKey: tls/server.key
        # Certificate governs the file location of the server TLS certificate.
        Certificate: tls/server.crt
        # RootCAs contains a list of additional root certificates used for verifying certificates
        # of other orderer nodes during outbound connections.
        # It is not required to be set, but can be used to augment the set of TLS CA certificates
        # available from the MSPs of each channel's configuration.
        RootCAs:
          - tls/ca.crt
        # Require client certificates / mutual TLS for inbound connections.
        ClientAuthRequired: false
        # If mutual TLS is enabled, ClientRootCAs contains a list of additional root certificates
        # used for verifying certificates of client connections.
        # It is not required to be set, but can be used to augment the set of TLS CA certificates
        # available from the MSPs of each channel's configuration.
        ClientRootCAs:
    # Keepalive settings for the GRPC server.
    Keepalive:
        # ServerMinInterval is the minimum permitted time between client pings.
        # If clients send pings more frequently, the server will
        # disconnect them.
        ServerMinInterval: 60s
        # ServerInterval is the time between pings to clients.
        ServerInterval: 7200s
        # ServerTimeout is the duration the server waits for a response from
        # a client before closing the connection.
        ServerTimeout: 20s

    # Since all nodes should be consistent it is recommended to keep
    # the default value of 100MB for MaxRecvMsgSize & MaxSendMsgSize
    # Max message size in bytes the GRPC server and client can receive
    MaxRecvMsgSize: 104857600
    # Max message size in bytes the GRPC server and client can send
    MaxSendMsgSize: 104857600

    # Cluster settings for ordering service nodes that communicate with other ordering service nodes
    # such as Raft based ordering service.
    Cluster:
        # SendBufferSize is the maximum number of messages in the egress buffer.
        # Consensus messages are dropped if the buffer is full, and transaction
        # messages are waiting for space to be freed.
        SendBufferSize: 100

        # ClientCertificate governs the file location of the client TLS certificate
        # used to establish mutual TLS connections with other ordering service nodes.
        # If not set, the server General.TLS.Certificate is re-used.
        ClientCertificate:
        # ClientPrivateKey governs the file location of the private key of the client TLS certificate.
        # If not set, the server General.TLS.PrivateKey is re-used.
        ClientPrivateKey:

        # The below 4 properties should be either set together, or be unset together.
        # If they are set, then the orderer node uses a separate listener for intra-cluster
        # communication. If they are unset, then the general orderer listener is used.
        # This is useful if you want to use a different TLS server certificates on the
        # client-facing and the intra-cluster listeners.

        # ListenPort defines the port on which the cluster listens to connections.
        ListenPort:
        # ListenAddress defines the IP on which to listen to intra-cluster communication.
        ListenAddress:
        # ServerCertificate defines the file location of the server TLS certificate used for intra-cluster
        # communication.
        ServerCertificate:
        # ServerPrivateKey defines the file location of the private key of the TLS certificate.
        ServerPrivateKey:

    # Bootstrap method: The method by which to obtain the bootstrap block
    # system channel is specified. The option can be one of:
    #   "file" - path to a file containing the genesis block or config block of system channel
    #   "none" - allows an orderer to start without a system channel configuration
    BootstrapMethod: file

    # Bootstrap file: The file containing the bootstrap block to use when
    # initializing the orderer system channel and BootstrapMethod is set to
    # "file".  The bootstrap file can be the genesis block, and it can also be
    # a config block for late bootstrap of some consensus methods like Raft.
    # Generate a genesis block by updating $FABRIC_CFG_PATH/configtx.yaml and
    # using configtxgen command with "-outputBlock" option.
    # Defaults to file "genesisblock" (in $FABRIC_CFG_PATH directory) if not specified.
    BootstrapFile:

    # LocalMSPDir is where to find the private crypto material needed by the
    # orderer. It is set relative here as a default for dev environments but
    # should be changed to the real location in production.
    LocalMSPDir: msp

    # LocalMSPID is the identity to register the local MSP material with the MSP
    # manager. IMPORTANT: The local MSP ID of an orderer needs to match the MSP
    # ID of one of the organizations defined in the orderer system channel's
    # /Channel/Orderer configuration. The sample organization defined in the
    # sample configuration provided has an MSP ID of "SampleOrg".
    LocalMSPID: SampleOrg

    # Enable an HTTP service for Go "pprof" profiling as documented at:
    # https://golang.org/pkg/net/http/pprof
    Profile:
        Enabled: false
        Address: 0.0.0.0:6060

    # BCCSP configures the blockchain crypto service providers.
    BCCSP:
        # Default specifies the preferred blockchain crypto service provider
        # to use. If the preferred provider is not available, the software
        # based provider ("SW") will be used.
        # Valid providers are:
        #  - SW: a software based crypto provider
        #  - PKCS11: a CA hardware security module crypto provider.
        Default: SW

        # SW configures the software based blockchain crypto provider.
        SW:
            # TODO: The default Hash and Security level needs refactoring to be
            # fully configurable. Changing these defaults requires coordination
            # SHA2 is hardcoded in several places, not only BCCSP
            Hash: SHA2
            Security: 256
            # Location of key store. If this is unset, a location will be
            # chosen using: 'LocalMSPDir'/keystore
            FileKeyStore:
                KeyStore:

        # Settings for the PKCS#11 crypto provider (i.e. when DEFAULT: PKCS11)
        PKCS11:
            # Location of the PKCS11 module library
            Library:
            # Token Label
            Label:
            # User PIN
            Pin:
            Hash:
            Security:
            FileKeyStore:
                KeyStore:

    # Authentication contains configuration parameters related to authenticating
    # client messages
    Authentication:
        # the acceptable difference between the current server time and the
        # client's time as specified in a client request message
        TimeWindow: 15m


################################################################################
#
#   SECTION: File Ledger
#
#   - This section applies to the configuration of the file ledger.
#
################################################################################
FileLedger:

    # Location: The directory to store the blocks in.
    Location: {{ .DataPath }}

################################################################################
#
#   Debug Configuration
#
#   - This controls the debugging options for the orderer
#
################################################################################
Debug:

    # BroadcastTraceDir when set will cause each request to the Broadcast service
    # for this orderer to be written to a file in this directory
    BroadcastTraceDir:

    # DeliverTraceDir when set will cause each request to the Deliver service
    # for this orderer to be written to a file in this directory
    DeliverTraceDir:

################################################################################
#
#   Operations Configuration
#
#   - This configures the operations server endpoint for the orderer
#
################################################################################
Operations:
    # host and port for the operations server
    ListenAddress: 127.0.0.1:8443

    # TLS configuration for the operations endpoint
    TLS:
        # TLS enabled
        Enabled: false

        # Certificate is the location of the PEM encoded TLS certificate
        Certificate:

        # PrivateKey points to the location of the PEM-encoded key
        PrivateKey:

        # Most operations service endpoints require client authentication when TLS
        # is enabled. ClientAuthRequired requires client certificate authentication
        # at the TLS layer to access all resources.
        ClientAuthRequired: false

        # Paths to PEM encoded ca certificates to trust for client authentication
        ClientRootCAs: []

################################################################################
#
#   Metrics Configuration
#
#   - This configures metrics collection for the orderer
#
################################################################################
Metrics:
    # The metrics provider is one of statsd, prometheus, or disabled
    Provider: disabled

    # The statsd configuration
    Statsd:
      # network type: tcp or udp
      Network: udp

      # the statsd server address
      Address: 127.0.0.1:8125

      # The interval at which locally cached counters and gauges are pushed
      # to statsd; timings are pushed immediately
      WriteInterval: 30s

      # The prefix is prepended to all emitted statsd metrics
      Prefix:

################################################################################
#
#   Admin Configuration
#
#   - This configures the admin server endpoint for the orderer
#
################################################################################
Admin:
    # host and port for the admin server
    ListenAddress: 127.0.0.1:9443

    # TLS configuration for the admin endpoint
    TLS:
        # TLS enabled
        Enabled: false

        # Certificate is the location of the PEM encoded TLS certificate
        Certificate:

        # PrivateKey points to the location of the PEM-encoded key
        PrivateKey:

        # Most admin service endpoints require client authentication when TLS
        # is enabled. ClientAuthRequired requires client certificate authentication
        # at the TLS layer to access all resources.
        #
        # NOTE: When TLS is enabled, the admin endpoint requires mutual TLS. The
        # orderer will panic on startup if this value is set to false.
        ClientAuthRequired: true

        # Paths to PEM encoded ca certificates to trust for client authentication
        ClientRootCAs: []

################################################################################
#
#   Channel participation API Configuration
#
#   - This provides the channel participation API configuration for the orderer.
#   - Channel participation uses the ListenAddress and TLS settings of the Admin
#     service.
#
################################################################################
ChannelParticipation:
    # Channel participation API is enabled.
    Enabled: false

    # The maximum size of the request body when joining a channel.
    MaxRequestBodySize: 1 MB


################################################################################
#
#   Consensus Configuration
#
#   - This section contains config options for a consensus plugin. It is opaque
#     to orderer, and completely up to consensus implementation to make use of.
#
################################################################################
Consensus:
    # The allowed key-value pairs here depend on consensus plugin. For etcd/raft,
    # we use following options:

    # WALDir specifies the location at which Write Ahead Logs for etcd/raft are
    # stored. Each channel will have its own subdir named after channel ID.
    WALDir: {{ .DataPath }}/etcdraft/wal

    # SnapDir specifies the location at which snapshots for etcd/raft are
    # stored. Each channel will have its own subdir named after channel ID.
    SnapDir: {{ .DataPath }}/etcdraft/snapshot

`

	data := struct {
		ListenAddress           string
		ListenPort              string
		OperationsListenAddress string
		AdminAddress            string
		DataPath                string
		MSPID                   string
	}{
		ListenAddress:           strings.Split(o.opts.ListenAddress, ":")[0],
		ListenPort:              strings.Split(o.opts.ListenAddress, ":")[1],
		OperationsListenAddress: o.opts.OperationsListenAddress,
		AdminAddress:            o.opts.AdminAddress,
		DataPath:                dataConfigPath,
		MSPID:                   o.mspID,
	}

	var buf bytes.Buffer
	tmpl := template.Must(template.New("orderer.yaml").Parse(ordererYamlTemplate))
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute orderer.yaml template: %w", err)
	}

	if err := os.WriteFile(filepath.Join(mspConfigPath, "orderer.yaml"), buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write orderer.yaml: %w", err)
	}

	return nil
}

// JoinChannel joins the orderer to a channel using the channel participation API
func (o *LocalOrderer) JoinChannel(genesisBlock []byte) error {
	ctx := context.Background()
	org, err := o.orgService.GetOrganization(ctx, o.org.ID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	adminTlsKeyDB, err := o.keyService.GetKey(ctx, int(org.AdminTlsKeyID.Int64))
	if err != nil {
		return fmt.Errorf("failed to get admin TLS key: %w", err)
	}
	adminTlsCert := adminTlsKeyDB.Certificate
	if adminTlsCert == nil {
		return fmt.Errorf("admin TLS certificate is nil")
	}
	if *adminTlsCert == "" {
		return fmt.Errorf("admin TLS certificate is empty")
	}
	adminTlsPK, err := o.keyService.GetDecryptedPrivateKey(int(org.AdminTlsKeyID.Int64))
	if err != nil {
		return fmt.Errorf("failed to get admin TLS private key: %w", err)
	}
	adminTlsCertX509, err := tls.X509KeyPair([]byte(*adminTlsCert), []byte(adminTlsPK))
	if err != nil {
		return fmt.Errorf("failed to get admin TLS certificate: %w", err)
	}
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM([]byte(org.TlsCertificate))
	if !ok {
		return fmt.Errorf("couldn't append certs")
	}
	ordererAdminUrl := fmt.Sprintf("https://%s", strings.Replace(o.opts.AdminAddress, "0.0.0.0", "127.0.0.1", 1))
	chResponse, err := osnadmin.Join(ordererAdminUrl, genesisBlock, certPool, adminTlsCertX509)
	if err != nil {
		return err
	}
	if chResponse.StatusCode == 405 {
		return fmt.Errorf("orderer already joined the channel")
	}
	responseData, err := io.ReadAll(chResponse.Body)
	if err != nil {
		return err
	}
	if chResponse.StatusCode != 201 {
		return fmt.Errorf("error joining orderer to channel: %d", chResponse.StatusCode)
	}

	var response osnadmin.ChannelInfo
	err = json.Unmarshal(responseData, &response)
	if err != nil {
		return err
	}

	return nil
}

// LeaveChannel removes the orderer from a channel using the channel participation API
func (o *LocalOrderer) LeaveChannel(channelID string) error {
	ctx := context.Background()
	// Get organization
	org, err := o.orgService.GetOrganization(ctx, o.organizationID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}
	tlsRootKeyDB, err := o.keyService.GetKey(ctx, int(org.TlsRootKeyID.Int64))
	if err != nil {
		return fmt.Errorf("failed to get TLS root key: %w", err)
	}
	tlsRootCert := tlsRootKeyDB.Certificate
	if tlsRootCert == nil {
		return fmt.Errorf("TLS root certificate is nil")
	}
	if *tlsRootCert == "" {
		return fmt.Errorf("TLS root certificate is empty")
	}

	tlsAdminKeyDB, err := o.keyService.GetKey(ctx, int(org.AdminTlsKeyID.Int64))
	if err != nil {
		return fmt.Errorf("failed to get TLS admin key: %w", err)
	}
	tlsAdminCert := tlsAdminKeyDB.Certificate
	if tlsAdminCert == nil {
		return fmt.Errorf("TLS admin certificate is nil")
	}
	if *tlsAdminCert == "" {
		return fmt.Errorf("TLS admin certificate is empty")
	}
	tlsAdminPK, err := o.keyService.GetDecryptedPrivateKey(int(org.AdminTlsKeyID.Int64))
	if err != nil {
		return fmt.Errorf("failed to get TLS admin private key: %w", err)
	}

	// Create CA cert pool
	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM([]byte(*tlsRootCert))
	if !ok {
		return fmt.Errorf("failed to append TLS root certificate to CA cert pool")
	}

	// Create client certificate
	cert, err := tls.X509KeyPair([]byte(*tlsAdminCert), []byte(tlsAdminPK))
	if err != nil {
		return fmt.Errorf("failed to load client certificate: %w", err)
	}
	adminAddress := strings.Replace(o.opts.AdminAddress, "0.0.0.0", "127.0.0.1", 1)
	// Call osnadmin Remove API
	resp, err := osnadmin.Remove(fmt.Sprintf("https://%s", adminAddress), channelID, caCertPool, cert)
	if err != nil {
		return fmt.Errorf("failed to remove orderer from channel: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to remove orderer from channel: status=%d, body=%s", resp.StatusCode, string(body))
	}

	o.logger.Info("Successfully removed orderer from channel", "orderer", o.opts.ID, "channel", channelID)
	return nil
}
