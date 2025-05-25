package chainlaunchdeploy

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/audit"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
)

// evmDeployer is a concrete implementation of the Deployer interface for EVM deployments.
type evmDeployer struct {
	auditService *audit.AuditService
}

// SetAuditService sets the audit service for logging deployment events
func (d *evmDeployer) SetAuditService(auditService *audit.AuditService) {
	d.auditService = auditService
}

// logAuditEvent is a helper to log audit events
func (d *evmDeployer) logAuditEvent(ctx context.Context, eventType string, outcome audit.EventOutcome, deploymentID string, details map[string]interface{}) {
	if d.auditService == nil {
		return
	}

	event := audit.NewEvent()
	event.EventSource = "chainlaunchdeploy"
	event.EventType = eventType
	event.EventOutcome = outcome
	event.AffectedResource = deploymentID
	event.Details = details
	event.Timestamp = time.Now().UTC()

	// Set severity based on outcome
	if outcome == audit.EventOutcomeFailure {
		event.Severity = audit.SeverityCritical
	} else {
		event.Severity = audit.SeverityInfo
	}

	// Use deployment ID as request ID if it's a valid UUID format (tx hash won't be)
	if _, err := uuid.Parse(deploymentID); err == nil {
		event.RequestID = uuid.MustParse(deploymentID)
	} else {
		event.RequestID = uuid.New()
	}

	// Log synchronously for compliance
	_ = d.auditService.LogEvent(ctx, event)
}

// DeployEVMContract deploys a smart contract to an EVM-compatible blockchain (e.g., Besu).
// It now accepts a DeploymentStatusReporter to track status updates.
func (d *evmDeployer) DeployEVMContract(params EVMParams, reporter DeploymentStatusReporter) (DeploymentResult, error) {
	ctx := context.Background()
	deploymentID := "evm-pending"

	// Log deployment start
	d.logAuditEvent(ctx, "EVM_DEPLOYMENT_START", audit.EventOutcomePending, deploymentID, map[string]interface{}{
		"chainID":              params.ChainID,
		"rpcURL":               params.RPCURL,
		"hasABI":               params.ABI != "",
		"hasBytecode":          len(params.Bytecode) > 0,
		"constructorArgsCount": len(params.ConstructorArgs),
	})

	reporter.ReportStatus(DeploymentStatusUpdate{
		DeploymentID: deploymentID,
		Status:       StatusPending,
		Message:      "Deployment pending",
	})

	if err := ValidateEVMParams(params); err != nil {
		d.logAuditEvent(ctx, "EVM_DEPLOYMENT_VALIDATION_FAILED", audit.EventOutcomeFailure, deploymentID, map[string]interface{}{
			"error": err.Error(),
		})
		reporter.ReportStatus(DeploymentStatusUpdate{
			DeploymentID: deploymentID,
			Status:       StatusFailed,
			Message:      "Validation failed",
			Error:        err,
		})
		return DeploymentResult{Success: false, Error: err}, err
	}
	if params.Signer == nil {
		err := errors.New("Signer function is required")
		d.logAuditEvent(ctx, "EVM_DEPLOYMENT_VALIDATION_FAILED", audit.EventOutcomeFailure, deploymentID, map[string]interface{}{
			"error": err.Error(),
		})
		reporter.ReportStatus(DeploymentStatusUpdate{
			DeploymentID: deploymentID,
			Status:       StatusFailed,
			Message:      "Signer function is required",
			Error:        err,
		})
		return DeploymentResult{Success: false, Error: err}, err
	}
	if params.ABI == "" {
		err := errors.New("ABI is required")
		d.logAuditEvent(ctx, "EVM_DEPLOYMENT_VALIDATION_FAILED", audit.EventOutcomeFailure, deploymentID, map[string]interface{}{
			"error": err.Error(),
		})
		reporter.ReportStatus(DeploymentStatusUpdate{
			DeploymentID: deploymentID,
			Status:       StatusFailed,
			Message:      "ABI is required",
			Error:        err,
		})
		return DeploymentResult{Success: false, Error: err}, err
	}
	if len(params.Bytecode) == 0 {
		err := errors.New("Bytecode is required")
		d.logAuditEvent(ctx, "EVM_DEPLOYMENT_VALIDATION_FAILED", audit.EventOutcomeFailure, deploymentID, map[string]interface{}{
			"error": err.Error(),
		})
		reporter.ReportStatus(DeploymentStatusUpdate{
			DeploymentID: deploymentID,
			Status:       StatusFailed,
			Message:      "Bytecode is required",
			Error:        err,
		})
		return DeploymentResult{Success: false, Error: err}, err
	}
	if params.RPCURL == "" {
		err := errors.New("RPCURL is required")
		d.logAuditEvent(ctx, "EVM_DEPLOYMENT_VALIDATION_FAILED", audit.EventOutcomeFailure, deploymentID, map[string]interface{}{
			"error": err.Error(),
		})
		reporter.ReportStatus(DeploymentStatusUpdate{
			DeploymentID: deploymentID,
			Status:       StatusFailed,
			Message:      "RPCURL is required",
			Error:        err,
		})
		return DeploymentResult{Success: false, Error: err}, err
	}
	if params.ChainID == 0 {
		err := errors.New("ChainID is required")
		d.logAuditEvent(ctx, "EVM_DEPLOYMENT_VALIDATION_FAILED", audit.EventOutcomeFailure, deploymentID, map[string]interface{}{
			"error": err.Error(),
		})
		reporter.ReportStatus(DeploymentStatusUpdate{
			DeploymentID: deploymentID,
			Status:       StatusFailed,
			Message:      "ChainID is required",
			Error:        err,
		})
		return DeploymentResult{Success: false, Error: err}, err
	}

	reporter.ReportStatus(DeploymentStatusUpdate{
		DeploymentID: deploymentID,
		Status:       StatusRunning,
		Message:      "Connecting to EVM node...",
	})

	client, err := ethclient.Dial(params.RPCURL)
	if err != nil {
		d.logAuditEvent(ctx, "EVM_DEPLOYMENT_CONNECTION_FAILED", audit.EventOutcomeFailure, deploymentID, map[string]interface{}{
			"error":  err.Error(),
			"rpcURL": params.RPCURL,
		})
		reporter.ReportStatus(DeploymentStatusUpdate{
			DeploymentID: deploymentID,
			Status:       StatusFailed,
			Message:      "Failed to connect to EVM node",
			Error:        err,
		})
		return DeploymentResult{Success: false, Error: err}, err
	}
	defer client.Close()

	parsedABI, err := abi.JSON(strings.NewReader(params.ABI))
	if err != nil {
		d.logAuditEvent(ctx, "EVM_DEPLOYMENT_ABI_PARSE_FAILED", audit.EventOutcomeFailure, deploymentID, map[string]interface{}{
			"error": err.Error(),
		})
		reporter.ReportStatus(DeploymentStatusUpdate{
			DeploymentID: deploymentID,
			Status:       StatusFailed,
			Message:      "Invalid ABI",
			Error:        err,
		})
		return DeploymentResult{Success: false, Error: fmt.Errorf("invalid ABI: %w", err)}, err
	}
	auth, err := bind.NewKeyedTransactorWithChainID(nil, big.NewInt(params.ChainID))
	if err != nil {
		d.logAuditEvent(ctx, "EVM_DEPLOYMENT_TRANSACTOR_FAILED", audit.EventOutcomeFailure, deploymentID, map[string]interface{}{
			"error":   err.Error(),
			"chainID": params.ChainID,
		})
		reporter.ReportStatus(DeploymentStatusUpdate{
			DeploymentID: deploymentID,
			Status:       StatusFailed,
			Message:      "Failed to create transactor",
			Error:        err,
		})
		return DeploymentResult{Success: false, Error: err}, err
	}
	auth.Signer = params.Signer

	reporter.ReportStatus(DeploymentStatusUpdate{
		DeploymentID: deploymentID,
		Status:       StatusRunning,
		Message:      "Deploying contract...",
	})

	address, tx, _, err := bind.DeployContract(auth, parsedABI, params.Bytecode, client, params.ConstructorArgs...)
	if err != nil {
		d.logAuditEvent(ctx, "EVM_DEPLOYMENT_FAILED", audit.EventOutcomeFailure, deploymentID, map[string]interface{}{
			"error":   err.Error(),
			"chainID": params.ChainID,
		})
		reporter.ReportStatus(DeploymentStatusUpdate{
			DeploymentID: deploymentID,
			Status:       StatusFailed,
			Message:      "Deployment failed",
			Error:        err,
		})
		return DeploymentResult{Success: false, Error: err}, err
	}

	deploymentID = tx.Hash().Hex() // Use tx hash as deployment ID from now on

	// Log transaction submitted
	d.logAuditEvent(ctx, "EVM_DEPLOYMENT_TX_SUBMITTED", audit.EventOutcomePending, deploymentID, map[string]interface{}{
		"transactionHash": tx.Hash().Hex(),
		"contractAddress": address.Hex(),
		"chainID":         params.ChainID,
		"gasPrice":        tx.GasPrice().String(),
		"gasLimit":        tx.Gas(),
	})

	reporter.ReportStatus(DeploymentStatusUpdate{
		DeploymentID: deploymentID,
		Status:       StatusRunning,
		Message:      "Waiting for transaction to be mined...",
	})

	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		d.logAuditEvent(ctx, "EVM_DEPLOYMENT_MINING_FAILED", audit.EventOutcomeFailure, deploymentID, map[string]interface{}{
			"error":           err.Error(),
			"transactionHash": tx.Hash().Hex(),
		})
		reporter.ReportStatus(DeploymentStatusUpdate{
			DeploymentID: deploymentID,
			Status:       StatusFailed,
			Message:      "Transaction mining failed",
			Error:        err,
		})
		return DeploymentResult{Success: false, Error: err}, err
	}
	if receipt.Status != 1 {
		err := fmt.Errorf("transaction failed: %s", tx.Hash().Hex())
		d.logAuditEvent(ctx, "EVM_DEPLOYMENT_TX_FAILED", audit.EventOutcomeFailure, deploymentID, map[string]interface{}{
			"transactionHash": tx.Hash().Hex(),
			"receiptStatus":   receipt.Status,
			"gasUsed":         receipt.GasUsed,
		})
		reporter.ReportStatus(DeploymentStatusUpdate{
			DeploymentID: deploymentID,
			Status:       StatusFailed,
			Message:      "Transaction failed",
			Error:        err,
		})
		return DeploymentResult{Success: false, Error: err}, err
	}

	// Log successful deployment
	d.logAuditEvent(ctx, "EVM_DEPLOYMENT_SUCCESS", audit.EventOutcomeSuccess, deploymentID, map[string]interface{}{
		"transactionHash":   tx.Hash().Hex(),
		"contractAddress":   address.Hex(),
		"chainID":           params.ChainID,
		"blockNumber":       receipt.BlockNumber.Uint64(),
		"gasUsed":           receipt.GasUsed,
		"cumulativeGasUsed": receipt.CumulativeGasUsed,
	})

	reporter.ReportStatus(DeploymentStatusUpdate{
		DeploymentID: deploymentID,
		Status:       StatusSuccess,
		Message:      fmt.Sprintf("Contract deployed at %s, tx: %s", address.Hex(), tx.Hash().Hex()),
	})

	return DeploymentResult{
		Success:         true,
		TransactionHash: tx.Hash().Hex(),
		ContractAddress: address.Hex(),
		Logs:            fmt.Sprintf("Contract deployed at %s, tx: %s", address.Hex(), tx.Hash().Hex()),
	}, nil
}

// NewDeployer returns a new instance of the EVM deployer for now.
func NewDeployer() Deployer {
	return &evmDeployer{}
}

// NewDeployerWithAudit returns a new instance of the EVM deployer with audit service
func NewDeployerWithAudit(auditService *audit.AuditService) DeployerWithAudit {
	return &evmDeployer{
		auditService: auditService,
	}
}
