package chainlaunchdeploy

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/audit"
	"github.com/google/uuid"
	"github.com/hyperledger/fabric-admin-sdk/pkg/chaincode"
)

// fabricDeployerAudit is a helper for audit logging in Fabric deployments
var fabricAuditService *audit.AuditService

// SetFabricAuditService sets the audit service for Fabric deployment logging
func SetFabricAuditService(auditService *audit.AuditService) {
	fabricAuditService = auditService
}

func logFabricAuditEvent(ctx context.Context, eventType string, outcome audit.EventOutcome, deploymentID string, details map[string]interface{}) {
	if fabricAuditService == nil {
		return
	}
	event := audit.NewEvent()
	event.EventSource = "chainlaunchdeploy"
	event.EventType = eventType
	event.EventOutcome = outcome
	event.AffectedResource = deploymentID
	event.Details = details
	event.Timestamp = time.Now().UTC()
	if outcome == audit.EventOutcomeFailure {
		event.Severity = audit.SeverityCritical
	} else {
		event.Severity = audit.SeverityInfo
	}
	if _, err := uuid.Parse(deploymentID); err == nil {
		event.RequestID = uuid.MustParse(deploymentID)
	} else {
		event.RequestID = uuid.New()
	}
	_ = fabricAuditService.LogEvent(ctx, event)
}

// InstallChaincode installs a chaincode package on a Fabric peer, reporting status at each stage.
func InstallChaincode(params FabricChaincodeInstallParams, reporter DeploymentStatusReporter) (DeploymentResult, error) {
	ctx := context.Background()
	deploymentID := fmt.Sprintf("install-%s", params.Label)
	logFabricAuditEvent(ctx, "FABRIC_CHAINCODE_INSTALL_START", audit.EventOutcomePending, deploymentID, map[string]interface{}{
		"label": params.Label,
	})
	reporter.ReportStatus(DeploymentStatusUpdate{
		DeploymentID: deploymentID,
		Status:       StatusPending,
		Message:      "Chaincode install pending",
	})
	if err := ValidateFabricChaincodeInstallParams(params); err != nil {
		logFabricAuditEvent(ctx, "FABRIC_CHAINCODE_INSTALL_VALIDATION_FAILED", audit.EventOutcomeFailure, deploymentID, map[string]interface{}{
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
	logFabricAuditEvent(ctx, "FABRIC_CHAINCODE_INSTALL_RUNNING", audit.EventOutcomePending, deploymentID, map[string]interface{}{
		"label": params.Label,
	})
	reporter.ReportStatus(DeploymentStatusUpdate{
		DeploymentID: deploymentID,
		Status:       StatusRunning,
		Message:      "Installing chaincode...",
	})
	// TODO: Implement install logic using fabric-admin-sdk
	logFabricAuditEvent(ctx, "FABRIC_CHAINCODE_INSTALL_SUCCESS", audit.EventOutcomeSuccess, deploymentID, map[string]interface{}{
		"label":  params.Label,
		"result": "Chaincode installed (stub)",
	})
	pkg := params.PackageBytes
	result, err := params.Peer.Install(ctx, bytes.NewReader(pkg))
	if err != nil {
		return DeploymentResult{Success: false, Error: err}, err
	}

	reporter.ReportStatus(DeploymentStatusUpdate{
		DeploymentID: deploymentID,
		Status:       StatusSuccess,
		Message:      fmt.Sprintf("Chaincode installed successfully %s", result.PackageId),
	})
	return DeploymentResult{Success: true, Logs: "Chaincode installed (stub)"}, nil
}

func ApproveChaincode(params FabricChaincodeApproveParams, reporter DeploymentStatusReporter) (DeploymentResult, error) {
	ctx := context.Background()
	deploymentID := fmt.Sprintf("approve-%s-%s", params.Name, params.Version)
	logFabricAuditEvent(ctx, "FABRIC_CHAINCODE_APPROVE_START", audit.EventOutcomePending, deploymentID, map[string]interface{}{
		"name":    params.Name,
		"version": params.Version,
	})
	reporter.ReportStatus(DeploymentStatusUpdate{
		DeploymentID: deploymentID,
		Status:       StatusPending,
		Message:      "Chaincode approve pending",
	})
	if err := ValidateFabricChaincodeApproveParams(params); err != nil {
		logFabricAuditEvent(ctx, "FABRIC_CHAINCODE_APPROVE_VALIDATION_FAILED", audit.EventOutcomeFailure, deploymentID, map[string]interface{}{
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
	logFabricAuditEvent(ctx, "FABRIC_CHAINCODE_APPROVE_RUNNING", audit.EventOutcomePending, deploymentID, map[string]interface{}{
		"name":    params.Name,
		"version": params.Version,
	})
	reporter.ReportStatus(DeploymentStatusUpdate{
		DeploymentID: deploymentID,
		Status:       StatusRunning,
		Message:      "Approving chaincode...",
	})
	applicationPolicy, err := chaincode.NewApplicationPolicy(params.EndorsementPolicy, "")
	if err != nil {
		logFabricAuditEvent(ctx, "FABRIC_CHAINCODE_APPROVE_VALIDATION_FAILED", audit.EventOutcomeFailure, deploymentID, map[string]interface{}{
			"error": err.Error(),
		})
		return DeploymentResult{Success: false, Error: err}, err
	}
	chaincodeDef := &chaincode.Definition{
		ChannelName:       params.ChannelID,
		PackageID:         params.PackageID,
		Name:              params.Name,
		Version:           params.Version,
		EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		Sequence:          int64(params.Sequence),
		InitRequired:      false,
		Collections:       nil,
		ApplicationPolicy: applicationPolicy,
	}
	err = params.Gateway.Approve(ctx, chaincodeDef)
	if err != nil {
		if strings.Contains(err.Error(), "redefine uncommitted") {
			logFabricAuditEvent(ctx, "FABRIC_CHAINCODE_APPROVE_SUCCESS", audit.EventOutcomeSuccess, deploymentID, map[string]interface{}{
				"name":    params.Name,
				"version": params.Version,
				"result":  "Chaincode approved (stub)",
			})
		} else {
			logFabricAuditEvent(ctx, "FABRIC_CHAINCODE_APPROVE_VALIDATION_FAILED", audit.EventOutcomeFailure, deploymentID, map[string]interface{}{
				"error": err.Error(),
			})
			return DeploymentResult{Success: false, Error: err}, err
		}
	}
	// TODO: Implement approve logic using fabric-admin-sdk
	logFabricAuditEvent(ctx, "FABRIC_CHAINCODE_APPROVE_SUCCESS", audit.EventOutcomeSuccess, deploymentID, map[string]interface{}{
		"name":    params.Name,
		"version": params.Version,
		"result":  "Chaincode approved (stub)",
	})
	reporter.ReportStatus(DeploymentStatusUpdate{
		DeploymentID: deploymentID,
		Status:       StatusSuccess,
		Message:      "Chaincode approved successfully",
	})
	return DeploymentResult{Success: true, Logs: "Chaincode approved (stub)"}, nil
}

func CommitChaincode(params FabricChaincodeCommitParams, reporter DeploymentStatusReporter) (DeploymentResult, error) {
	ctx := context.Background()
	deploymentID := fmt.Sprintf("commit-%s-%s", params.Name, params.Version)
	logFabricAuditEvent(ctx, "FABRIC_CHAINCODE_COMMIT_START", audit.EventOutcomePending, deploymentID, map[string]interface{}{
		"name":    params.Name,
		"version": params.Version,
	})
	reporter.ReportStatus(DeploymentStatusUpdate{
		DeploymentID: deploymentID,
		Status:       StatusPending,
		Message:      "Chaincode commit pending",
	})
	if err := ValidateFabricChaincodeCommitParams(params); err != nil {
		logFabricAuditEvent(ctx, "FABRIC_CHAINCODE_COMMIT_VALIDATION_FAILED", audit.EventOutcomeFailure, deploymentID, map[string]interface{}{
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
	logFabricAuditEvent(ctx, "FABRIC_CHAINCODE_COMMIT_RUNNING", audit.EventOutcomePending, deploymentID, map[string]interface{}{
		"name":    params.Name,
		"version": params.Version,
	})
	reporter.ReportStatus(DeploymentStatusUpdate{
		DeploymentID: deploymentID,
		Status:       StatusRunning,
		Message:      "Committing chaincode...",
	})
	applicationPolicy, err := chaincode.NewApplicationPolicy(params.EndorsementPolicy, "")
	if err != nil {
		logFabricAuditEvent(ctx, "FABRIC_CHAINCODE_COMMIT_VALIDATION_FAILED", audit.EventOutcomeFailure, deploymentID, map[string]interface{}{
			"error": err.Error(),
		})
		return DeploymentResult{Success: false, Error: err}, err
	}

	chaincodeDef := &chaincode.Definition{
		ChannelName:       params.ChannelID,
		Name:              params.Name,
		Version:           params.Version,
		EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		Sequence:          int64(params.Sequence),
		InitRequired:      false,
		Collections:       nil,
		ApplicationPolicy: applicationPolicy,
	}
	err = params.Gateway.Commit(ctx, chaincodeDef)
	if err != nil {
		logFabricAuditEvent(ctx, "FABRIC_CHAINCODE_COMMIT_VALIDATION_FAILED", audit.EventOutcomeFailure, deploymentID, map[string]interface{}{
			"error": err.Error(),
		})
		return DeploymentResult{Success: false, Error: err}, err
	}
	// TODO: Implement commit logic using fabric-admin-sdk
	logFabricAuditEvent(ctx, "FABRIC_CHAINCODE_COMMIT_SUCCESS", audit.EventOutcomeSuccess, deploymentID, map[string]interface{}{
		"name":    params.Name,
		"version": params.Version,
		"result":  "Chaincode committed (stub)",
	})
	reporter.ReportStatus(DeploymentStatusUpdate{
		DeploymentID: deploymentID,
		Status:       StatusSuccess,
		Message:      "Chaincode committed successfully",
	})
	return DeploymentResult{Success: true, Logs: "Chaincode committed (stub)"}, nil
}

func DeployChaincode(params FabricChaincodeDeployParams, reporter DeploymentStatusReporter) (DeploymentResult, error) {
	ctx := context.Background()
	deploymentID := fmt.Sprintf("deploy-%s-%s", params.ApproveParams.Name, params.ApproveParams.Version)
	logFabricAuditEvent(ctx, "FABRIC_CHAINCODE_DEPLOY_START", audit.EventOutcomePending, deploymentID, map[string]interface{}{
		"name":    params.ApproveParams.Name,
		"version": params.ApproveParams.Version,
	})
	reporter.ReportStatus(DeploymentStatusUpdate{
		DeploymentID: deploymentID,
		Status:       StatusPending,
		Message:      "Chaincode deploy pending",
	})
	if err := ValidateFabricChaincodeInstallParams(params.InstallParams); err != nil {
		logFabricAuditEvent(ctx, "FABRIC_CHAINCODE_DEPLOY_INSTALL_VALIDATION_FAILED", audit.EventOutcomeFailure, deploymentID, map[string]interface{}{
			"error": err.Error(),
		})
		reporter.ReportStatus(DeploymentStatusUpdate{
			DeploymentID: deploymentID,
			Status:       StatusFailed,
			Message:      "Install validation failed",
			Error:        err,
		})
		return DeploymentResult{Success: false, Error: err}, err
	}
	if err := ValidateFabricChaincodeApproveParams(params.ApproveParams); err != nil {
		logFabricAuditEvent(ctx, "FABRIC_CHAINCODE_DEPLOY_APPROVE_VALIDATION_FAILED", audit.EventOutcomeFailure, deploymentID, map[string]interface{}{
			"error": err.Error(),
		})
		reporter.ReportStatus(DeploymentStatusUpdate{
			DeploymentID: deploymentID,
			Status:       StatusFailed,
			Message:      "Approve validation failed",
			Error:        err,
		})
		return DeploymentResult{Success: false, Error: err}, err
	}
	if err := ValidateFabricChaincodeCommitParams(params.CommitParams); err != nil {
		logFabricAuditEvent(ctx, "FABRIC_CHAINCODE_DEPLOY_COMMIT_VALIDATION_FAILED", audit.EventOutcomeFailure, deploymentID, map[string]interface{}{
			"error": err.Error(),
		})
		reporter.ReportStatus(DeploymentStatusUpdate{
			DeploymentID: deploymentID,
			Status:       StatusFailed,
			Message:      "Commit validation failed",
			Error:        err,
		})
		return DeploymentResult{Success: false, Error: err}, err
	}
	logFabricAuditEvent(ctx, "FABRIC_CHAINCODE_DEPLOY_RUNNING", audit.EventOutcomePending, deploymentID, map[string]interface{}{
		"name":    params.ApproveParams.Name,
		"version": params.ApproveParams.Version,
	})
	reporter.ReportStatus(DeploymentStatusUpdate{
		DeploymentID: deploymentID,
		Status:       StatusRunning,
		Message:      "Deploying chaincode (install, approve, commit)...",
	})
	// TODO: Implement deploy logic using fabric-admin-sdk
	logFabricAuditEvent(ctx, "FABRIC_CHAINCODE_DEPLOY_SUCCESS", audit.EventOutcomeSuccess, deploymentID, map[string]interface{}{
		"name":    params.ApproveParams.Name,
		"version": params.ApproveParams.Version,
		"result":  "Chaincode deployed (stub)",
	})
	reporter.ReportStatus(DeploymentStatusUpdate{
		DeploymentID: deploymentID,
		Status:       StatusSuccess,
		Message:      "Chaincode deployed successfully",
	})
	return DeploymentResult{Success: true, Logs: "Chaincode deployed (stub)"}, nil
}

type fabricDeployer struct{}

// NewFabricDeployer returns a new instance of the Fabric deployer
func NewFabricDeployer() Deployer {
	return &fabricDeployer{}
}

// Install installs a chaincode package on a Fabric peer
func (d *fabricDeployer) Install(params FabricChaincodeInstallParams, reporter DeploymentStatusReporter) (DeploymentResult, error) {
	return InstallChaincode(params, reporter)
}

// Approve approves a chaincode definition for an organization
func (d *fabricDeployer) Approve(params FabricChaincodeApproveParams, reporter DeploymentStatusReporter) (DeploymentResult, error) {
	return ApproveChaincode(params, reporter)
}

// Commit commits a chaincode definition to the channel
func (d *fabricDeployer) Commit(params FabricChaincodeCommitParams, reporter DeploymentStatusReporter) (DeploymentResult, error) {
	return CommitChaincode(params, reporter)
}

// DeployFabricContract implements the Deployer interface for Fabric
func (d *fabricDeployer) DeployFabricContract(params FabricChaincodeDeployParams, reporter DeploymentStatusReporter) (DeploymentResult, error) {
	installResult, err := d.Install(params.InstallParams, reporter)
	if err != nil {
		return installResult, err
	}
	approveResult, err := d.Approve(params.ApproveParams, reporter)
	if err != nil {
		return approveResult, err
	}
	commitResult, err := d.Commit(params.CommitParams, reporter)
	if err != nil {
		return commitResult, err
	}
	return commitResult, nil
}

// DeployEVMContract is not supported for fabricDeployer
func (d *fabricDeployer) DeployEVMContract(params EVMParams, reporter DeploymentStatusReporter) (DeploymentResult, error) {
	return DeploymentResult{Success: false, Error: fmt.Errorf("DeployEVMContract not supported for Fabric")}, fmt.Errorf("DeployEVMContract not supported for Fabric")
}
