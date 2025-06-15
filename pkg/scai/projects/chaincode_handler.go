package projects

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/chainlaunch/chainlaunch/pkg/http/response"
	"github.com/go-chi/chi/v5"
)

// HandlerRequest represents the request structure for both invoke and query operations
type HandlerRequest struct {
	Function string   `json:"function" example:"createAsset" description:"Name of the chaincode function to invoke"`
	Args     []string `json:"args" example:"[\"asset1\",\"100\"]" description:"Array of arguments to pass to the function"`
	OrgID    int64    `json:"orgId" example:"1" description:"ID of the organization that will sign the transaction"`
	KeyID    int64    `json:"keyId" example:"1" description:"ID of the key to use for signing the transaction"`
}

// HandlerResponse represents the response structure for both invoke and query operations
type HandlerResponse struct {
	Status    string      `json:"status"`
	Message   string      `json:"message"`
	Project   string      `json:"project"`
	Function  string      `json:"function"`
	Args      []string    `json:"args"`
	Result    interface{} `json:"result"`
	Channel   string      `json:"channel"`
	Chaincode string      `json:"chaincode"`
}

// @Summary Invoke a chaincode transaction
// @Description Invokes a transaction on the specified chaincode project
// @Tags Chaincode Projects
// @Accept json
// @Produce json
// @Param id path int true "Chaincode Project ID"
// @Param request body HandlerRequest true "Transaction parameters"
// @Success 200 {object} HandlerResponse "Transaction result"
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 404 {object} response.ErrorResponse "Project not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /chaincode-projects/{id}/invoke [post]
func (h *ProjectsHandler) InvokeTransaction(w http.ResponseWriter, r *http.Request) error {
	projectIDStr := chi.URLParam(r, "id")
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid project ID: %w", err)
	}

	var req HandlerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}

	// Convert handler request to service request
	serviceReq := TransactionRequest{
		ProjectID: projectID,
		Function:  req.Function,
		Args:      req.Args,
		OrgID:     req.OrgID,
		KeyID:     req.KeyID,
	}

	result, err := h.chaincodeService.InvokeTransaction(r.Context(), serviceReq)
	if err != nil {
		return fmt.Errorf("failed to invoke transaction: %w", err)
	}

	// Convert service response to handler response
	handlerResp := HandlerResponse{
		Status:    result.Status,
		Message:   result.Message,
		Project:   result.Project,
		Function:  result.Function,
		Args:      result.Args,
		Result:    result.Result,
		Channel:   result.Channel,
		Chaincode: result.Chaincode,
	}

	response.JSON(w, http.StatusOK, handlerResp)
	return nil
}

// @Summary Query a chaincode transaction
// @Description Queries the state of the specified chaincode project
// @Tags Chaincode Projects
// @Accept json
// @Produce json
// @Param id path int true "Chaincode Project ID"
// @Param request body HandlerRequest true "Query parameters"
// @Success 200 {object} HandlerResponse "Query result"
// @Failure 400 {object} response.ErrorResponse "Invalid request"
// @Failure 404 {object} response.ErrorResponse "Project not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /chaincode-projects/{id}/query [post]
func (h *ProjectsHandler) QueryTransaction(w http.ResponseWriter, r *http.Request) error {
	projectIDStr := chi.URLParam(r, "id")
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid project ID: %w", err)
	}

	var req HandlerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}

	// Convert handler request to service request
	serviceReq := TransactionRequest{
		ProjectID: projectID,
		Function:  req.Function,
		Args:      req.Args,
		OrgID:     req.OrgID,
		KeyID:     req.KeyID,
	}

	result, err := h.chaincodeService.QueryTransaction(r.Context(), serviceReq)
	if err != nil {
		return fmt.Errorf("failed to query transaction: %w", err)
	}

	// Convert service response to handler response
	handlerResp := HandlerResponse{
		Status:    result.Status,
		Message:   result.Message,
		Project:   result.Project,
		Function:  result.Function,
		Args:      result.Args,
		Result:    result.Result,
		Channel:   result.Channel,
		Chaincode: result.Chaincode,
	}

	response.JSON(w, http.StatusOK, handlerResp)
	return nil
}

// func (h *ChaincodeHandler) RegisterRoutes(r chi.Router) {
// 	r.Route("/chaincode-projects", func(r chi.Router) {
// 		r.Post("/{id}/invoke", response.Middleware(h.InvokeTransaction))
// 		r.Post("/{id}/query", response.Middleware(h.QueryTransaction))
// 	})
// }
