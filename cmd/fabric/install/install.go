package install

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/pkg/errors"

	"io"
	"strings"
	"time"

	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/lifecycle"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/policydsl"
	"github.com/spf13/cobra"
)

type installCmd struct {
	chaincode        string
	channel          string
	networkConfig    string
	users            []string
	organizations    []string
	signaturePolicy  string
	chaincodeAddress string
	envFile          string
	metaInfPath      string
	pdcFile          string
	local            bool
	rootCert         string
	clientCert       string
	clientKey        string
	logger           *logger.Logger
}

func (c installCmd) start() error {
	var chaincodeEndpoint string

	if c.local {
		// Use local chaincode address directly
		chaincodeEndpoint = c.chaincodeAddress
	} else {
		return errors.New("a tunnel is not supported in this version")
	}

	label := c.chaincode
	codeTarBytes, err := c.getCodeTarGz(
		chaincodeEndpoint,
		c.rootCert,
		c.clientKey,
		c.clientCert,
		c.metaInfPath,
	)
	if err != nil {
		return err
	}
	pkg, err := c.getChaincodePackage(c.chaincode, codeTarBytes)
	if err != nil {
		return err
	}
	_ = pkg
	packageID := lifecycle.ComputePackageID(label, pkg)
	c.logger.Infof("packageID: %s", packageID)

	// install chaincode in peers
	configBackend := config.FromFile(c.networkConfig)

	clientsMap := map[string]*resmgmt.Client{}
	sdk, err := fabsdk.New(configBackend)
	if err != nil {
		return err
	}
	for idx, mspID := range c.organizations {
		clientContext := sdk.Context(fabsdk.WithUser(c.users[idx]), fabsdk.WithOrg(mspID))
		clientsMap[mspID], err = resmgmt.New(clientContext)
		if err != nil {
			return err
		}
	}
	for mspID, resmgmtClient := range clientsMap {
		_, err = resmgmtClient.LifecycleInstallCC(
			resmgmt.LifecycleInstallCCRequest{
				Label:   label,
				Package: pkg,
			},
			resmgmt.WithTimeout(fab.ResMgmt, 20*time.Minute),
			resmgmt.WithTimeout(fab.PeerResponse, 20*time.Minute),
		)
		if err != nil {
			return err
		}
		c.logger.Infof("Chaincode installed in %s", mspID)
	}
	sp, err := policydsl.FromString(c.signaturePolicy)
	if err != nil {
		return err
	}

	version := "1"
	sequence := 1
	resmgmtClient := clientsMap[c.organizations[0]]
	committedCCs, err := resmgmtClient.LifecycleQueryCommittedCC(
		c.channel,
		resmgmt.LifecycleQueryCommittedCCRequest{Name: c.chaincode},
		resmgmt.WithTargetFilter(&multipleMSPFilter{mspIDs: c.organizations}),
	)
	if err != nil {
		c.logger.Warnf("Error when getting commited chaincodes: %v", err)
	}

	var collections []*pb.CollectionConfig
	if c.pdcFile != "" {
		//
		pdcBytes, err := ioutil.ReadFile(c.pdcFile)
		if err != nil {
			return err
		}
		collections, err = getCollectionConfigFromBytes([]byte(pdcBytes))
		if err != nil {
			return err
		}
	}
	c.logger.Infof("Commited CCs=%d", len(committedCCs))
	shouldCommit := len(committedCCs) == 0
	if len(committedCCs) > 0 {
		firstCommittedCC := committedCCs[0]
		signaturePolicyString := firstCommittedCC.SignaturePolicy.String()
		newSignaturePolicyString := sp.String()
		if signaturePolicyString != newSignaturePolicyString {
			c.logger.Infof("Signature policy changed, old=%s new=%s", signaturePolicyString, newSignaturePolicyString)
			shouldCommit = true
		} else {
			c.logger.Infof("Signature policy not changed, signaturePolicy=%s", signaturePolicyString)
		}
		// compare collections
		oldCollections := firstCommittedCC.CollectionConfig
		newCollections := collections
		if len(oldCollections) != len(newCollections) {
			c.logger.Infof("Collection config changed, old=%d new=%d", len(oldCollections), len(newCollections))
			shouldCommit = true
		} else {
			for idx, oldCollection := range oldCollections {
				oldCollectionPayload := oldCollection.Payload.(*pb.CollectionConfig_StaticCollectionConfig)
				newCollection := newCollections[idx]
				newCollectionPayload := newCollection.Payload.(*pb.CollectionConfig_StaticCollectionConfig)
				if oldCollectionPayload.StaticCollectionConfig.Name != newCollectionPayload.StaticCollectionConfig.Name {
					c.logger.Infof("Collection config changed, old=%s new=%s", oldCollectionPayload.StaticCollectionConfig.Name, newCollectionPayload.StaticCollectionConfig.Name)
					shouldCommit = true
					break
				}
				oldCollectionPolicy := oldCollection.GetStaticCollectionConfig().MemberOrgsPolicy
				newCollectionPolicy := newCollection.GetStaticCollectionConfig().MemberOrgsPolicy
				if oldCollectionPolicy.GetSignaturePolicy().String() != newCollectionPolicy.GetSignaturePolicy().String() {
					c.logger.Infof("Collection config changed, old=%s new=%s", oldCollectionPolicy.GetSignaturePolicy().String(), newCollectionPolicy.GetSignaturePolicy().String())
					shouldCommit = true
					break
				}
			}
		}
	}
	if len(committedCCs) > 0 {
		if shouldCommit {
			version = committedCCs[len(committedCCs)-1].Version
			sequence = int(committedCCs[len(committedCCs)-1].Sequence) + 1
		} else {
			version = committedCCs[len(committedCCs)-1].Version
			sequence = int(committedCCs[len(committedCCs)-1].Sequence)
		}
		c.logger.Infof("Chaincode already committed, version=%s sequence=%d", version, sequence)
	}
	c.logger.Infof("Should commit=%v", shouldCommit)
	// approve chaincode in orgs
	approveCCRequest := resmgmt.LifecycleApproveCCRequest{
		Name:              label,
		Version:           version,
		PackageID:         packageID,
		Sequence:          int64(sequence),
		CollectionConfig:  collections,
		EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		SignaturePolicy:   sp,
		InitRequired:      false,
	}
	for mspID, resmgmtClient := range clientsMap {

		txID, err := resmgmtClient.LifecycleApproveCC(
			c.channel,
			approveCCRequest,
			resmgmt.WithTargetFilter(&mspFilter{mspID: mspID}),
			resmgmt.WithTimeout(fab.ResMgmt, 20*time.Minute),
			resmgmt.WithTimeout(fab.PeerResponse, 20*time.Minute),
		)
		if err != nil && !strings.Contains(err.Error(), "redefine uncommitted") {
			c.logger.Errorf("Error when approving chaincode: %v", err)
			return err
		}
		c.logger.Infof("Chaincode approved, org=%s tx=%s", mspID, txID)
	}
	if shouldCommit {
		// commit chaincode in orgs
		txID, err := resmgmtClient.LifecycleCommitCC(
			c.channel,
			resmgmt.LifecycleCommitCCRequest{
				Name:              label,
				Version:           version,
				Sequence:          int64(sequence),
				CollectionConfig:  collections,
				EndorsementPlugin: "escc",
				ValidationPlugin:  "vscc",
				SignaturePolicy:   sp,
				InitRequired:      false,
			},
			resmgmt.WithTimeout(fab.ResMgmt, 2*time.Minute),
			resmgmt.WithTimeout(fab.PeerResponse, 2*time.Minute),
			resmgmt.WithTargetFilter(&multipleMSPFilter{mspIDs: c.organizations}),
		)
		if err != nil {
			return err
		}
		c.logger.Infof("Chaincode committed, tx=%s", txID)
	}
	sdk.Close()

	if c.envFile != "" {
		err = os.WriteFile(c.envFile, []byte(fmt.Sprintf(`
CORE_CHAINCODE_ADDRESS=%s
CORE_CHAINCODE_ID=%s
CORE_PEER_TLS_ENABLED=false
`, c.chaincodeAddress, packageID)), 0777)
		if err != nil {
			c.logger.Warn("Failed to write .env file: %s", err)
			return err
		}
	}
	return nil
}

type multipleMSPFilter struct {
	mspIDs []string
}

// Accept returns true if this peer is to be included in the target list
func (f *multipleMSPFilter) Accept(peer fab.Peer) bool {
	// check if its of one of the mspIDs
	for _, mspID := range f.mspIDs {
		if peer.MSPID() == mspID {
			return true
		}
	}
	return false
}

type mspFilter struct {
	mspID string
}

// Accept returns true if this peer is to be included in the target list
func (f *mspFilter) Accept(peer fab.Peer) bool {
	return peer.MSPID() == f.mspID
}

func (c *installCmd) getChaincodePackage(label string, codeTarGz []byte) ([]byte, error) {
	var err error
	metadataJson := fmt.Sprintf(`
{
  "type": "ccaas",
  "label": "%s"
}
`, label)
	// set up the output file
	buf := &bytes.Buffer{}

	// set up the gzip writer
	gw := gzip.NewWriter(buf)
	defer func(gw *gzip.Writer) {
		err := gw.Close()
		if err != nil {
			c.logger.Warnf("gzip.Writer.Close() failed: %s", err)
		}
	}(gw)
	tw := tar.NewWriter(gw)
	defer func(tw *tar.Writer) {
		err := tw.Close()
		if err != nil {
			c.logger.Warnf("tar.Writer.Close() failed: %s", err)
		}
	}(tw)
	header := new(tar.Header)
	header.Name = "metadata.json"
	metadataJsonBytes := []byte(metadataJson)
	header.Size = int64(len(metadataJsonBytes))
	header.Mode = 0777
	err = tw.WriteHeader(header)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(metadataJsonBytes)
	_, err = io.Copy(tw, r)
	if err != nil {
		return nil, err
	}
	headerCode := new(tar.Header)
	headerCode.Name = "code.tar.gz"
	headerCode.Size = int64(len(codeTarGz))
	headerCode.Mode = 0777
	err = tw.WriteHeader(headerCode)
	if err != nil {
		return nil, err
	}
	r = bytes.NewReader(codeTarGz)
	_, err = io.Copy(tw, r)
	if err != nil {
		return nil, err
	}
	err = tw.Close()
	if err != nil {
		return nil, err
	}
	err = gw.Close()
	if err != nil {
		c.logger.Warnf("gzip.Writer.Close() failed: %s", err)
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *installCmd) getCodeTarGz(
	address string,
	rootCert string,
	clientKey string,
	clientCert string,
	metaInfPath string,
) ([]byte, error) {
	var err error
	// Determine if TLS is required based on certificate presence
	tlsRequired := rootCert != ""
	clientAuthRequired := clientCert != "" && clientKey != ""

	// Read certificate files if provided
	var rootCertContent, clientKeyContent, clientCertContent string
	if tlsRequired {
		rootCertBytes, err := os.ReadFile(rootCert)
		if err != nil {
			return nil, fmt.Errorf("failed to read root certificate: %w", err)
		}
		rootCertContent = string(rootCertBytes)
	}

	if clientAuthRequired {
		clientKeyBytes, err := os.ReadFile(clientKey)
		if err != nil {
			return nil, fmt.Errorf("failed to read client key: %w", err)
		}
		clientKeyContent = string(clientKeyBytes)

		clientCertBytes, err := os.ReadFile(clientCert)
		if err != nil {
			return nil, fmt.Errorf("failed to read client certificate: %w", err)
		}
		clientCertContent = string(clientCertBytes)
	}

	connMap := map[string]interface{}{
		"address":              address,
		"dial_timeout":         "10s",
		"tls_required":         tlsRequired,
		"root_cert":            rootCertContent,
		"client_auth_required": clientAuthRequired,
		"client_key":           clientKeyContent,
		"client_cert":          clientCertContent,
	}
	connJsonBytes, err := json.Marshal(connMap)
	if err != nil {
		return nil, err
	}
	c.logger.Debugf("Conn=%s", string(connJsonBytes))
	// set up the output file
	buf := &bytes.Buffer{}
	// set up the gzip writer
	gw := gzip.NewWriter(buf)
	tw := tar.NewWriter(gw)
	header := new(tar.Header)
	header.Name = "connection.json"
	header.Size = int64(len(connJsonBytes))
	header.Mode = 0755
	err = tw.WriteHeader(header)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(connJsonBytes)
	_, err = io.Copy(tw, r)
	if err != nil {
		return nil, err
	}
	if metaInfPath != "" {
		src := metaInfPath
		// walk through 3 file in the folder
		err = filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
			// generate tar header
			header, err := tar.FileInfoHeader(fi, file)
			if err != nil {
				return err
			}

			// must provide real name
			// (see https://golang.org/src/archive/tar/common.go?#L626)
			relname, err := filepath.Rel(src, file)
			if err != nil {
				return err
			}
			if relname == "." {
				return nil
			}
			header.Name = "META-INF/" + filepath.ToSlash(relname)

			// write header
			if err := tw.WriteHeader(header); err != nil {
				return err
			}
			// if not a dir, write file content
			if !fi.IsDir() {
				data, err := os.Open(file)
				if err != nil {
					return err
				}
				if _, err := io.Copy(tw, data); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	err = tw.Close()
	if err != nil {
		return nil, err
	}
	err = gw.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func NewInstallCmd(logger *logger.Logger) *cobra.Command {
	c := &installCmd{
		logger: logger,
	}
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install the chaincode",
		Long:  `Install the chaincode`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c.logger.Infof("Installing the chaincode")

			return c.start()
		},
	}

	f := cmd.Flags()
	f.StringVar(&c.chaincode, "chaincode", "", "chaincode name within the channel")
	f.StringVar(&c.channel, "channel", "", "Channel name")
	f.StringVar(&c.networkConfig, "config", "", "Network config file")
	f.StringVar(&c.signaturePolicy, "policy", "", "Signature policy for the chaincode")
	f.StringArrayVarP(&c.organizations, "organizations", "o", []string{}, "Organizations to connect to ")
	f.StringArrayVarP(&c.users, "users", "u", []string{}, "Users to use")
	f.StringVar(&c.chaincodeAddress, "chaincodeAddress", "", "address of the local chaincode server, example: localhost:9999")
	f.StringVar(&c.envFile, "envFile", "", ".env file to write the environments variables")
	f.StringVar(&c.pdcFile, "pdc", "", "pdc file json, see examples/pdc.json")
	f.StringVar(&c.metaInfPath, "metaInf", "", "metadata")
	f.StringVar(&c.rootCert, "rootCert", "", "path to the root certificate file")
	f.StringVar(&c.clientCert, "clientCert", "", "path to the client certificate file")
	f.StringVar(&c.clientKey, "clientKey", "", "path to the client key file")
	f.BoolVar(&c.local, "local", false, "Use local chaincode address without ngrok tunnel")
	return cmd
}

type endorsementPolicy struct {
	ChannelConfigPolicy string `json:"channelConfigPolicy,omitempty"`
	SignaturePolicy     string `json:"signaturePolicy,omitempty"`
}

type collectionConfigJson struct {
	Name              string             `json:"name"`
	Policy            string             `json:"policy"`
	RequiredPeerCount *int32             `json:"requiredPeerCount"`
	MaxPeerCount      *int32             `json:"maxPeerCount"`
	BlockToLive       uint64             `json:"blockToLive"`
	MemberOnlyRead    bool               `json:"memberOnlyRead"`
	MemberOnlyWrite   bool               `json:"memberOnlyWrite"`
	EndorsementPolicy *endorsementPolicy `json:"endorsementPolicy,omitempty"`
}

// getCollectionConfig retrieves the collection configuration
// from the supplied byte array; the byte array must contain a
// json-formatted array of collectionConfigJson elements
func getCollectionConfigFromBytes(cconfBytes []byte) ([]*pb.CollectionConfig, error) {
	cconf := &[]collectionConfigJson{}
	err := json.Unmarshal(cconfBytes, cconf)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse the collection configuration")
	}

	ccarray := make([]*pb.CollectionConfig, 0, len(*cconf))
	for _, cconfitem := range *cconf {
		p, err := policydsl.FromString(cconfitem.Policy)
		if err != nil {
			return nil, errors.WithMessagef(err, "invalid policy %s", cconfitem.Policy)
		}

		cpc := &pb.CollectionPolicyConfig{
			Payload: &pb.CollectionPolicyConfig_SignaturePolicy{
				SignaturePolicy: p,
			},
		}

		var ep *pb.ApplicationPolicy
		if cconfitem.EndorsementPolicy != nil {
			signaturePolicy := cconfitem.EndorsementPolicy.SignaturePolicy
			channelConfigPolicy := cconfitem.EndorsementPolicy.ChannelConfigPolicy
			ep, err = getApplicationPolicy(signaturePolicy, channelConfigPolicy)
			if err != nil {
				return nil, errors.WithMessagef(err, "invalid endorsement policy [%#v]", cconfitem.EndorsementPolicy)
			}
		}

		// Set default requiredPeerCount and MaxPeerCount if not specified in json
		requiredPeerCount := int32(0)
		maxPeerCount := int32(1)
		if cconfitem.RequiredPeerCount != nil {
			requiredPeerCount = *cconfitem.RequiredPeerCount
		}
		if cconfitem.MaxPeerCount != nil {
			maxPeerCount = *cconfitem.MaxPeerCount
		}

		cc := &pb.CollectionConfig{
			Payload: &pb.CollectionConfig_StaticCollectionConfig{
				StaticCollectionConfig: &pb.StaticCollectionConfig{
					Name:              cconfitem.Name,
					MemberOrgsPolicy:  cpc,
					RequiredPeerCount: requiredPeerCount,
					MaximumPeerCount:  maxPeerCount,
					BlockToLive:       cconfitem.BlockToLive,
					MemberOnlyRead:    cconfitem.MemberOnlyRead,
					MemberOnlyWrite:   cconfitem.MemberOnlyWrite,
					EndorsementPolicy: ep,
				},
			},
		}

		ccarray = append(ccarray, cc)
	}

	return ccarray, nil
}

func getApplicationPolicy(signaturePolicy, channelConfigPolicy string) (*pb.ApplicationPolicy, error) {
	if signaturePolicy == "" && channelConfigPolicy == "" {
		// no policy, no problem
		return nil, nil
	}

	if signaturePolicy != "" && channelConfigPolicy != "" {
		// mo policies, mo problems
		return nil, errors.New(`cannot specify both "--signature-policy" and "--channel-config-policy"`)
	}

	var applicationPolicy *pb.ApplicationPolicy
	if signaturePolicy != "" {
		signaturePolicyEnvelope, err := policydsl.FromString(signaturePolicy)
		if err != nil {
			return nil, errors.Errorf("invalid signature policy: %s", signaturePolicy)
		}

		applicationPolicy = &pb.ApplicationPolicy{
			Type: &pb.ApplicationPolicy_SignaturePolicy{
				SignaturePolicy: signaturePolicyEnvelope,
			},
		}
	}

	if channelConfigPolicy != "" {
		applicationPolicy = &pb.ApplicationPolicy{
			Type: &pb.ApplicationPolicy_ChannelConfigPolicyReference{
				ChannelConfigPolicyReference: channelConfigPolicy,
			},
		}
	}

	return applicationPolicy, nil
}
