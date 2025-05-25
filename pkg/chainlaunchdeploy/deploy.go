package chainlaunchdeploy

import (
	"github.com/chainlaunch/chainlaunch/pkg/audit"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

// EVMParams defines the parameters required for EVM (e.g., Besu) smart contract deployment.
type EVMParams struct {
	SolidityCode    string        // (Optional) Solidity source code (for reference)
	ABI             string        // Contract ABI (JSON string)
	Bytecode        []byte        // Compiled contract bytecode
	RPCURL          string        // RPC endpoint for Besu node
	ChainID         int64         // Chain ID for the target network
	ConstructorArgs []interface{} // Constructor arguments for the contract
	Signer          bind.SignerFn // Signer function to sign transactions (delegated to caller for security)
	// Add more fields as needed (e.g., gas, nonce, etc.)
}

// FabricChaincodeInstallParams defines parameters for chaincode installation.
type FabricChaincodeInstallParams struct {
	PackageBytes []byte // Chaincode package bytes
	Label        string // Chaincode label
}

// FabricChaincodeApproveParams defines parameters for chaincode approval.
type FabricChaincodeApproveParams struct {
	Name              string
	Version           string
	Sequence          int64
	PackageID         string
	ChannelID         string
	EndorsementPolicy string
	CollectionsConfig []byte // Serialized CollectionConfigPackage
	InitRequired      bool
}

// FabricChaincodeCommitParams defines parameters for chaincode commit.
type FabricChaincodeCommitParams struct {
	Name              string
	Version           string
	Sequence          int64
	ChannelID         string
	EndorsementPolicy string
	CollectionsConfig []byte // Serialized CollectionConfigPackage
	InitRequired      bool
}

// FabricChaincodeDeployParams defines parameters for chaincode deployment (install+approve+commit).
type FabricChaincodeDeployParams struct {
	InstallParams FabricChaincodeInstallParams
	ApproveParams FabricChaincodeApproveParams
	CommitParams  FabricChaincodeCommitParams
}

// DeploymentResult represents the result of a deployment operation.
type DeploymentResult struct {
	Success         bool
	TransactionHash string // For EVM
	ContractAddress string // For EVM
	ChaincodeID     string // For Fabric
	Logs            string
	Error           error
}

// Deployer defines the public interface for the deployment module.
type Deployer interface {
	DeployEVMContract(params EVMParams, reporter DeploymentStatusReporter) (DeploymentResult, error)
}

// DeployerWithAudit extends Deployer with audit logging capability
type DeployerWithAudit interface {
	Deployer
	SetAuditService(auditService *audit.AuditService)
}

// Validation stubs for Fabric operations
func ValidateFabricChaincodeInstallParams(params FabricChaincodeInstallParams) error {
	if len(params.PackageBytes) == 0 {
		return ErrMissingField("PackageBytes")
	}
	if params.Label == "" {
		return ErrMissingField("Label")
	}
	return nil
}

func ValidateFabricChaincodeApproveParams(params FabricChaincodeApproveParams) error {
	if params.Name == "" {
		return ErrMissingField("Name")
	}
	if params.Version == "" {
		return ErrMissingField("Version")
	}
	if params.Sequence <= 0 {
		return ErrMissingField("Sequence")
	}
	if params.PackageID == "" {
		return ErrMissingField("PackageID")
	}
	if params.ChannelID == "" {
		return ErrMissingField("ChannelID")
	}
	return nil
}

func ValidateFabricChaincodeCommitParams(params FabricChaincodeCommitParams) error {
	if params.Name == "" {
		return ErrMissingField("Name")
	}
	if params.Version == "" {
		return ErrMissingField("Version")
	}
	if params.Sequence <= 0 {
		return ErrMissingField("Sequence")
	}
	if params.ChannelID == "" {
		return ErrMissingField("ChannelID")
	}
	return nil
}

// ErrMissingField is a helper for validation errors.
func ErrMissingField(field string) error {
	return &ValidationError{Field: field, Message: field + " is required"}
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
