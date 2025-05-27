package chainlaunchdeploy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/chainlaunch/chainlaunch/pkg/audit"
	"github.com/chainlaunch/chainlaunch/pkg/db"
	"github.com/chainlaunch/chainlaunch/pkg/errors"
	"github.com/chainlaunch/chainlaunch/pkg/http/response"
	"github.com/chainlaunch/chainlaunch/pkg/logger"
	nodeService "github.com/chainlaunch/chainlaunch/pkg/nodes/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for smart contract deployment
type Handler struct {
	auditService     *audit.AuditService
	logger           *logger.Logger
	besuDeployer     DeployerWithAudit
	validate         *validator.Validate
	nodeService      *nodeService.NodeService
	chaincodeService *ChaincodeService
}

// NewHandler creates a new smart contract deploy handler
func NewHandler(auditService *audit.AuditService, logger *logger.Logger, besuDeployer DeployerWithAudit, nodeService *nodeService.NodeService, chaincodeService *ChaincodeService) *Handler {
	SetFabricAuditService(auditService)
	if besuDeployer != nil {
		besuDeployer.SetAuditService(auditService)
	}
	return &Handler{
		auditService:     auditService,
		logger:           logger,
		besuDeployer:     besuDeployer,
		validate:         validator.New(),
		nodeService:      nodeService,
		chaincodeService: chaincodeService,
	}
}

// RegisterRoutes registers the smart contract deploy routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/sc/fabric", func(r chi.Router) {
		r.Post("/deploy", response.Middleware(h.DeployFabricChaincode))
		r.Post("/peer/{peerId}/chaincode/install", response.Middleware(h.InstallFabricChaincode))
		r.Post("/peer/{peerId}/chaincode/approve", response.Middleware(h.ApproveFabricChaincode))
		r.Post("/peer/{peerId}/chaincode/commit", response.Middleware(h.CommitFabricChaincode))
		r.Post("/docker-deploy", response.Middleware(h.DeployFabricChaincodeWithDockerImage))
		r.Get("/chaincodes", response.Middleware(h.ListFabricChaincodes))
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

// FabricInstallRequest represents the request body for Fabric chaincode install
// (separate from service struct for HTTP layer)
type FabricInstallRequest struct {
	PackageBytes []byte `json:"package_bytes" validate:"required"`
	Label        string `json:"label" validate:"required"`
}

type FabricInstallResponse struct {
	Status  string           `json:"status"`
	Message string           `json:"message"`
	Result  DeploymentResult `json:"result"`
}

// FabricApproveRequest represents the request body for Fabric chaincode approve
type FabricApproveRequest struct {
	Name              string
	Version           string
	Sequence          int64
	PackageID         string
	ChannelID         string
	EndorsementPolicy string
	InitRequired      bool
}

type FabricApproveResponse struct {
	Status  string           `json:"status"`
	Message string           `json:"message"`
	Result  DeploymentResult `json:"result"`
}

// FabricCommitRequest represents the request body for Fabric chaincode commit
type FabricCommitRequest struct {
	Name              string
	Version           string
	Sequence          int64
	ChannelID         string
	EndorsementPolicy string
	InitRequired      bool
}

type FabricCommitResponse struct {
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

// FabricChaincodeDockerDeployRequest represents the request body for Fabric chaincode Docker deployment
// (separate from service struct for HTTP layer)
type FabricChaincodeDockerDeployRequest struct {
	Name          string `json:"name" validate:"required"`
	Slug          string `json:"slug"` // optional, for updates
	DockerImage   string `json:"docker_image" validate:"required"`
	PackageID     string `json:"package_id" validate:"required"`
	HostPort      string `json:"host_port"`      // optional, if empty a free port is chosen
	ContainerPort string `json:"container_port"` // optional, defaults to 7052
}

type FabricChaincodeDockerDeployResponse struct {
	Status  string           `json:"status"`
	Message string           `json:"message"`
	Slug    string           `json:"slug"`
	Result  DeploymentResult `json:"result"`
}

type ListFabricChaincodesResponse struct {
	Chaincodes []db.FabricChaincode `json:"chaincodes"`
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

// InstallFabricChaincode handles Fabric chaincode install requests
// @Summary Install Fabric chaincode
// @Description Install a chaincode package on a Fabric peer
// @Tags SmartContracts
// @Accept json
// @Produce json
// @Param peerId path string true "Peer ID"
// @Param request body FabricInstallRequest true "Fabric chaincode install parameters"
// @Success 200 {object} FabricInstallResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /sc/fabric/peer/{peerId}/chaincode/install [post]
func (h *Handler) InstallFabricChaincode(w http.ResponseWriter, r *http.Request) error {
	peerId := chi.URLParam(r, "peerId")
	peerIdInt, err := strconv.ParseInt(peerId, 10, 64)
	if err != nil {
		h.logger.Error("Invalid peer ID", "peerId", peerId)
		return errors.NewValidationError("invalid peer ID", map[string]interface{}{
			"detail": "Invalid peer ID",
			"code":   "INVALID_PEER_ID",
		})
	}
	var req FabricInstallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid Fabric install request body", "error", err)
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
	peerService, err := h.nodeService.GetFabricPeerService(r.Context(), peerIdInt)
	if err != nil {
		h.logger.Error("Node not found", "peerId", peerId)
		return errors.NewValidationError("node not found", map[string]interface{}{
			"detail": "Node not found",
			"code":   "NODE_NOT_FOUND",
		})
	}

	params := FabricChaincodeInstallParams{
		Peer:         peerService,
		PackageBytes: req.PackageBytes, Label: req.Label,
	}
	reporter := NewInMemoryDeploymentStatusReporter()
	result, err := InstallChaincode(params, reporter)
	if err != nil {
		h.logger.Error("Fabric chaincode install failed", "error", err)
		return errors.NewInternalError("install failed", err, nil)
	}
	resp := FabricInstallResponse{
		Status:  "success",
		Message: "Chaincode installed successfully",
		Result:  result,
	}
	return response.WriteJSON(w, http.StatusOK, resp)
}

// ApproveFabricChaincode handles Fabric chaincode approve requests
// @Summary Approve Fabric chaincode
// @Description Approve a chaincode definition for an organization
// @Tags SmartContracts
// @Accept json
// @Produce json
// @Param peerId path string true "Peer ID"
// @Param request body FabricApproveRequest true "Fabric chaincode approve parameters"
// @Success 200 {object} FabricApproveResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /sc/fabric/peer/{peerId}/chaincode/approve [post]
func (h *Handler) ApproveFabricChaincode(w http.ResponseWriter, r *http.Request) error {
	peerId := chi.URLParam(r, "peerId")
	var req FabricApproveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid Fabric approve request body", "error", err)
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
	peerIdInt, err := strconv.ParseInt(peerId, 10, 64)
	if err != nil {
		h.logger.Error("Invalid peer ID", "peerId", peerId)
		return errors.NewValidationError("invalid peer ID", map[string]interface{}{
			"detail": "Invalid peer ID",
			"code":   "INVALID_PEER_ID",
		})
	}
	peerGateway, err := h.nodeService.GetFabricPeerGateway(r.Context(), peerIdInt)
	if err != nil {
		h.logger.Error("Node not found", "peerId", peerId)
		return errors.NewValidationError("node not found", map[string]interface{}{
			"detail": "Node not found",
			"code":   "NODE_NOT_FOUND",
		})
	}

	params := FabricChaincodeApproveParams{
		Gateway:           peerGateway,
		Name:              req.Name,
		Version:           req.Version,
		Sequence:          req.Sequence,
		PackageID:         req.PackageID,
		ChannelID:         req.ChannelID,
		EndorsementPolicy: req.EndorsementPolicy,
		InitRequired:      req.InitRequired,
	}
	reporter := NewInMemoryDeploymentStatusReporter()
	result, err := ApproveChaincode(params, reporter)
	if err != nil {
		h.logger.Error("Fabric chaincode approve failed", "error", err)
		return errors.NewInternalError("approve failed", err, nil)
	}
	resp := FabricApproveResponse{
		Status:  "success",
		Message: "Chaincode approved successfully",
		Result:  result,
	}
	return response.WriteJSON(w, http.StatusOK, resp)
}

// CommitFabricChaincode handles Fabric chaincode commit requests
// @Summary Commit Fabric chaincode
// @Description Commit a chaincode definition to the channel
// @Tags SmartContracts
// @Accept json
// @Produce json
// @Param peerId path string true "Peer ID"
// @Param request body FabricCommitRequest true "Fabric chaincode commit parameters"
// @Success 200 {object} FabricCommitResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /sc/fabric/peer/{peerId}/chaincode/commit [post]
func (h *Handler) CommitFabricChaincode(w http.ResponseWriter, r *http.Request) error {
	peerId := chi.URLParam(r, "peerId")
	var req FabricCommitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid Fabric commit request body", "error", err)
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
	peerIdInt, err := strconv.ParseInt(peerId, 10, 64)
	if err != nil {
		h.logger.Error("Invalid peer ID", "peerId", peerId)
		return errors.NewValidationError("invalid peer ID", map[string]interface{}{
			"detail": "Invalid peer ID",
			"code":   "INVALID_PEER_ID",
		})
	}
	peerGateway, err := h.nodeService.GetFabricPeerGateway(r.Context(), peerIdInt)
	if err != nil {
		h.logger.Error("Node not found", "peerId", peerId)
		return errors.NewValidationError("node not found", map[string]interface{}{
			"detail": "Node not found",
			"code":   "NODE_NOT_FOUND",
		})
	}
	params := FabricChaincodeCommitParams{
		Gateway:           peerGateway,
		Name:              req.Name,
		Version:           req.Version,
		Sequence:          req.Sequence,
		ChannelID:         req.ChannelID,
		EndorsementPolicy: req.EndorsementPolicy,
		InitRequired:      req.InitRequired,
	}
	reporter := NewInMemoryDeploymentStatusReporter()
	result, err := CommitChaincode(params, reporter)
	if err != nil {
		h.logger.Error("Fabric chaincode commit failed", "error", err)
		return errors.NewInternalError("commit failed", err, nil)
	}
	resp := FabricCommitResponse{
		Status:  "success",
		Message: "Chaincode committed successfully",
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

// ListFabricChaincodes lists all deployed Fabric chaincodes
// @Summary List deployed Fabric chaincodes
// @Description List all Fabric chaincodes deployed via Docker
// @Tags SmartContracts
// @Accept json
// @Produce json
// @Success 200 {object} ListFabricChaincodesResponse
// @Failure 500 {object} response.Response
// @Router /sc/fabric/chaincodes [get]
func (h *Handler) ListFabricChaincodes(w http.ResponseWriter, r *http.Request) error {
	chaincodes, err := h.chaincodeService.ListChaincodes(r.Context())
	if err != nil {
		h.logger.Error("Failed to list fabric chaincodes", "error", err)
		return errors.NewInternalError("failed to list chaincodes", err, nil)
	}
	resp := ListFabricChaincodesResponse{Chaincodes: make([]db.FabricChaincode, 0, len(chaincodes))}
	for _, cc := range chaincodes {
		if cc != nil {
			resp.Chaincodes = append(resp.Chaincodes, *cc)
		}
	}
	return response.WriteJSON(w, http.StatusOK, resp)
}

// DeployFabricChaincodeWithDockerImage handles Fabric chaincode Docker deployment requests
// @Summary Deploy Fabric chaincode with Docker image
// @Description Deploy a chaincode to a Fabric network using a Docker image, package ID, and port mapping. If host_port is empty, a free port is chosen. If container_port is empty, defaults to 7052.
// @Tags SmartContracts
// @Accept json
// @Produce json
// @Param request body FabricChaincodeDockerDeployRequest true "Fabric chaincode Docker deployment parameters (host_port: optional, container_port: optional, defaults to 7052)"
// @Success 200 {object} FabricChaincodeDockerDeployResponse
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /sc/fabric/docker-deploy [post]
func (h *Handler) DeployFabricChaincodeWithDockerImage(w http.ResponseWriter, r *http.Request) error {
	var req FabricChaincodeDockerDeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Invalid Fabric Docker deploy request body", "error", err)
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
	result, err := DeployChaincodeWithDockerImage(req.DockerImage, req.PackageID, req.HostPort, req.ContainerPort, reporter)
	if err != nil {
		h.logger.Error("Fabric chaincode Docker deployment failed", "error", err)
		return errors.NewInternalError("docker deployment failed", err, nil)
	}

	slug := req.Slug
	if slug == "" {
		slug = generateUniqueSlug(req.Name)
	}

	// Try to update if slug exists, else insert
	chaincode, err := h.chaincodeService.GetChaincodeBySlug(r.Context(), slug)
	if err == nil && chaincode != nil && chaincode.ID > 0 {
		chaincode, err = h.chaincodeService.UpdateChaincodeBySlug(r.Context(), slug, req.DockerImage, req.PackageID, req.HostPort, req.ContainerPort, "running")
		if err != nil {
			h.logger.Error("Failed to update fabric chaincode record", "error", err)
			return errors.NewInternalError("failed to update chaincode record", err, nil)
		}
	} else {
		chaincode, err = h.chaincodeService.InsertChaincode(r.Context(), req.Name, slug, req.PackageID, req.DockerImage, req.HostPort, req.ContainerPort, "running")
		if err != nil {
			h.logger.Error("Failed to insert fabric chaincode record", "error", err)
			return errors.NewInternalError("failed to insert chaincode record", err, nil)
		}
	}

	resp := FabricChaincodeDockerDeployResponse{
		Status:  "success",
		Message: "Chaincode Docker container started successfully",
		Slug:    slug,
		Result:  result,
	}
	return response.WriteJSON(w, http.StatusOK, resp)
}

// generateUniqueSlug creates a slug from the name and a random suffix if needed
func generateUniqueSlug(name string) string {
	base := strings.ToLower(regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(name, "-"))
	return fmt.Sprintf("%s-%s", strings.Trim(base, "-"), uuid.New().String()[:8])
}
