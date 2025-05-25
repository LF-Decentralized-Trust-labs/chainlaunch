package chainlaunchdeploy

import (
	"encoding/json"
	"net/http"

	"github.com/chainlaunch/chainlaunch/pkg/audit"
	"github.com/chainlaunch/chainlaunch/pkg/errors"
	"github.com/chainlaunch/chainlaunch/pkg/http/response"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// Handler handles HTTP requests for smart contract deployment
type Handler struct {
	auditService *audit.AuditService
	logger       *logger.Logger
	besuDeployer DeployerWithAudit
	validate     *validator.Validate
}

// NewHandler creates a new smart contract deploy handler
func NewHandler(auditService *audit.AuditService, logger *logger.Logger, besuDeployer DeployerWithAudit) *Handler {
	SetFabricAuditService(auditService)
	if besuDeployer != nil {
		besuDeployer.SetAuditService(auditService)
	}
	return &Handler{
		auditService: auditService,
		logger:       logger,
		besuDeployer: besuDeployer,
		validate:     validator.New(),
	}
}

// RegisterRoutes registers the smart contract deploy routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/sc/fabric", func(r chi.Router) {
		r.Post("/deploy", response.Middleware(h.DeployFabricChaincode))
	})
	r.Route("/sc/besu", func(r chi.Router) {
		r.Post("/deploy", response.Middleware(h.DeployBesuContract))
	})
}

// FabricDeployRequest represents the request body for Fabric chaincode deployment
type FabricDeployRequest FabricChaincodeDeployParams

// FabricDeployResponse represents the response for Fabric chaincode deployment
type FabricDeployResponse struct {
	Status  string           `json:"status"`
	Message string           `json:"message"`
	Result  DeploymentResult `json:"result"`
}

// BesuDeployRequest represents the request body for Besu contract deployment
type BesuDeployRequest EVMParams

// BesuDeployResponse represents the response for Besu contract deployment
type BesuDeployResponse struct {
	Status  string           `json:"status"`
	Message string           `json:"message"`
	Result  DeploymentResult `json:"result"`
}

// DeployFabricChaincode handles Fabric chaincode deployment requests
// @Summary Deploy Fabric chaincode
// @Description Deploy a chaincode to a Fabric network (install, approve, commit)
// @Tags SmartContracts
// @Accept json
// @Produce json
// @Param request body FabricDeployRequest true "Fabric chaincode deployment parameters"
// @Success 200 {object} FabricDeployResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /sc/fabric/deploy [post]
func (h *Handler) DeployFabricChaincode(w http.ResponseWriter, r *http.Request) error {
	var req FabricDeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid Fabric request body", "error", err)
		return errors.NewValidationError("invalid request body", map[string]interface{}{
			"detail": err.Error(),
			"code":   "INVALID_REQUEST_BODY",
		})
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors[err.Field()] = err.Tag()
		}
		return errors.NewValidationError("validation failed", map[string]interface{}{
			"detail": "Request validation failed",
			"code":   "VALIDATION_ERROR",
			"errors": validationErrors,
		})
	}

	reporter := NewInMemoryDeploymentStatusReporter()
	result, err := DeployChaincode(FabricChaincodeDeployParams(req), reporter)
	if err != nil {
		h.logger.Error("Fabric chaincode deployment failed", "error", err)
		return errors.NewInternalError("deployment failed", err, nil)
	}

	resp := FabricDeployResponse{
		Status:  "success",
		Message: "Chaincode deployed successfully",
		Result:  result,
	}
	return response.WriteJSON(w, http.StatusOK, resp)
}

// DeployBesuContract handles Besu contract deployment requests
// @Summary Deploy Besu smart contract
// @Description Deploy a smart contract to a Besu (EVM) network
// @Tags SmartContracts
// @Accept json
// @Produce json
// @Param request body BesuDeployRequest true "Besu contract deployment parameters"
// @Success 200 {object} BesuDeployResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /sc/besu/deploy [post]
func (h *Handler) DeployBesuContract(w http.ResponseWriter, r *http.Request) error {
	var req BesuDeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid Besu request body", "error", err)
		return errors.NewValidationError("invalid request body", map[string]interface{}{
			"detail": err.Error(),
			"code":   "INVALID_REQUEST_BODY",
		})
	}

	if h.besuDeployer == nil {
		return errors.NewInternalError("Besu deployer not configured", nil, nil)
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := make(map[string]string)
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors[err.Field()] = err.Tag()
		}
		return errors.NewValidationError("validation failed", map[string]interface{}{
			"detail": "Request validation failed",
			"code":   "VALIDATION_ERROR",
			"errors": validationErrors,
		})
	}

	reporter := NewInMemoryDeploymentStatusReporter()
	result, err := h.besuDeployer.DeployEVMContract(EVMParams(req), reporter)
	if err != nil {
		h.logger.Error("Besu contract deployment failed", "error", err)
		return errors.NewInternalError("deployment failed", err, nil)
	}

	resp := BesuDeployResponse{
		Status:  "success",
		Message: "Contract deployed successfully",
		Result:  result,
	}
	return response.WriteJSON(w, http.StatusOK, resp)
}
