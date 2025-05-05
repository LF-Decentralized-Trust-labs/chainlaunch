package common

import (
	"encoding/json"
	"fmt"
	stdhttp "net/http"

	"github.com/chainlaunch/chainlaunch/pkg/nodes/types"
)

// PaginatedNodesResponse represents a paginated list of nodes
type PaginatedNodesResponse struct {
	Items       []NodeResponse `json:"items"`
	Total       int64          `json:"total"`
	Page        int            `json:"page"`
	PageCount   int            `json:"pageCount"`
	HasNextPage bool           `json:"hasNextPage"`
}

// NodeResponse represents a node response
type NodeResponse struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	NodeType     string `json:"nodeType"`
	Status       string `json:"status"`
	Endpoint     string `json:"endpoint"`
	ErrorMessage string `json:"errorMessage"`
}

// CreatePeerNode creates a new Fabric peer node
func (c *Client) CreatePeerNode(req *types.FabricPeerConfig) (*NodeResponse, error) {
	body := map[string]interface{}{
		"name":               req.Name,
		"blockchainPlatform": "FABRIC",
		"fabricPeer":         req,
	}

	resp, err := c.Post("/nodes", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer node: %w", err)
	}

	if err := CheckResponse(resp, stdhttp.StatusCreated); err != nil {
		return nil, fmt.Errorf("failed to create peer node: %w", err)
	}

	var node NodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &node, nil
}

// CreateOrdererNode creates a new Fabric orderer node
func (c *Client) CreateOrdererNode(req *types.FabricOrdererConfig) (*NodeResponse, error) {
	body := map[string]interface{}{
		"name":               req.Name,
		"blockchainPlatform": "FABRIC",
		"fabricOrderer":      req,
	}

	resp, err := c.Post("/nodes", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create orderer node: %w", err)
	}

	if err := CheckResponse(resp, stdhttp.StatusCreated); err != nil {
		return nil, fmt.Errorf("failed to create orderer node: %w", err)
	}

	var node NodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &node, nil
}

// ListNodes lists all nodes with optional platform filter
func (c *Client) ListNodes(platform string, page, limit int) (*PaginatedNodesResponse, error) {
	path := "/nodes"
	if platform != "" {
		path = fmt.Sprintf("/nodes/platform/%s", platform)
	}

	if page > 0 {
		path = fmt.Sprintf("%s?page=%d", path, page)
		if limit > 0 {
			path = fmt.Sprintf("%s&limit=%d", path, limit)
		}
	}

	resp, err := c.Get(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	if err := CheckResponse(resp, stdhttp.StatusOK); err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	var nodes PaginatedNodesResponse
	if err := json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &nodes, nil
}

// ListPeerNodes lists all Fabric peer nodes
func (c *Client) ListPeerNodes(page, limit int) (*PaginatedNodesResponse, error) {
	nodes, err := c.ListNodes("FABRIC", page, limit)
	if err != nil {
		return nil, err
	}

	// Filter only peer nodes
	var peerNodes PaginatedNodesResponse
	for _, node := range nodes.Items {
		if node.NodeType == "FABRIC_PEER" {
			peerNodes.Items = append(peerNodes.Items, node)
		}
	}
	peerNodes.Total = int64(len(peerNodes.Items))
	peerNodes.Page = nodes.Page
	peerNodes.PageCount = nodes.PageCount
	peerNodes.HasNextPage = nodes.HasNextPage

	return &peerNodes, nil
}

// ListOrdererNodes lists all Fabric orderer nodes
func (c *Client) ListOrdererNodes(page, limit int) (*PaginatedNodesResponse, error) {
	nodes, err := c.ListNodes("FABRIC", page, limit)
	if err != nil {
		return nil, err
	}

	// Filter only orderer nodes
	var ordererNodes PaginatedNodesResponse
	for _, node := range nodes.Items {
		if node.NodeType == "FABRIC_ORDERER" {
			ordererNodes.Items = append(ordererNodes.Items, node)
		}
	}
	ordererNodes.Total = int64(len(ordererNodes.Items))
	ordererNodes.Page = nodes.Page
	ordererNodes.PageCount = nodes.PageCount
	ordererNodes.HasNextPage = nodes.HasNextPage

	return &ordererNodes, nil
}

// DeleteNode deletes a node by ID
func (c *Client) DeleteNode(id int64) error {
	resp, err := c.Delete(fmt.Sprintf("/nodes/%d", id))
	if err != nil {
		return fmt.Errorf("failed to delete node: %w", err)
	}

	if err := CheckResponse(resp, stdhttp.StatusNoContent); err != nil {
		return fmt.Errorf("failed to delete node: %w", err)
	}

	return nil
}

// UpdatePeerNode updates a Fabric peer node
func (c *Client) UpdatePeerNode(id int64, req *types.FabricPeerConfig) (*NodeResponse, error) {
	body := map[string]interface{}{
		"blockchainPlatform": "FABRIC",
		"fabricPeer":         req,
	}

	resp, err := c.Put(fmt.Sprintf("/nodes/%d", id), body)
	if err != nil {
		return nil, fmt.Errorf("failed to update peer node: %w", err)
	}

	if err := CheckResponse(resp, stdhttp.StatusOK); err != nil {
		return nil, fmt.Errorf("failed to update peer node: %w", err)
	}

	var node NodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &node, nil
}

// UpdateOrdererNode updates a Fabric orderer node
func (c *Client) UpdateOrdererNode(id int64, req *types.FabricOrdererConfig) (*NodeResponse, error) {
	body := map[string]interface{}{
		"blockchainPlatform": "FABRIC",
		"fabricOrderer":      req,
	}

	resp, err := c.Put(fmt.Sprintf("/nodes/%d", id), body)
	if err != nil {
		return nil, fmt.Errorf("failed to update orderer node: %w", err)
	}

	if err := CheckResponse(resp, stdhttp.StatusOK); err != nil {
		return nil, fmt.Errorf("failed to update orderer node: %w", err)
	}

	var node NodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &node, nil
}
