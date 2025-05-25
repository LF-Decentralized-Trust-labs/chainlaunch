package chainlaunchdeploy

import (
	"context"
	"fmt"
	"time"

	"github.com/chainlaunch/chainlaunch/pkg/audit"
	"github.com/google/uuid"
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
	reporter.ReportStatus(DeploymentStatusUpdate{
		DeploymentID: deploymentID,
		Status:       StatusSuccess,
		Message:      "Chaincode installed successfully",
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
