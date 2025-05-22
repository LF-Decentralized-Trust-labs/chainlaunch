package common

import (
	"encoding/json"
	"fmt"
	stdhttp "net/http"

	httptypes "github.com/chainlaunch/chainlaunch/pkg/nodes/http"
	"github.com/chainlaunch/chainlaunch/pkg/nodes/types"
)

// CreatePeerNode creates a new Fabric peer node
func (c *Client) CreatePeerNode(req *types.FabricPeerConfig) (*httptypes.NodeResponse, error) {
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

	var node httptypes.NodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &node, nil
}

// CreateOrdererNode creates a new Fabric orderer node
func (c *Client) CreateOrdererNode(req *types.FabricOrdererConfig) (*httptypes.NodeResponse, error) {
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

	var node httptypes.NodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &node, nil
}

// CreateBesuNode creates a new Besu node
func (c *Client) CreateBesuNode(name string, req *types.BesuNodeConfig) (*httptypes.NodeResponse, error) {
	body := map[string]interface{}{
		"name":               name,
		"blockchainPlatform": "BESU",
		"besuNode":           req,
	}

	resp, err := c.Post("/nodes", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create besu node: %w", err)
	}

	if err := CheckResponse(resp, stdhttp.StatusCreated); err != nil {
		return nil, fmt.Errorf("failed to create besu node: %w", err)
	}

	var node httptypes.NodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &node, nil
}

// ListNodes lists all nodes with optional platform filter
func (c *Client) ListNodes(platform string, page, limit int) (*httptypes.ListNodesResponse, error) {
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

	var nodes httptypes.ListNodesResponse
	if err := json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &nodes, nil
}

// ListPeerNodes lists all Fabric peer nodes
func (c *Client) ListPeerNodes(page, limit int) (*httptypes.ListNodesResponse, error) {
	nodes, err := c.ListNodes("FABRIC", page, limit)
	if err != nil {
		return nil, err
	}

	// Filter only peer nodes
	var peerNodes httptypes.ListNodesResponse
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
func (c *Client) ListOrdererNodes(page, limit int) (*httptypes.ListNodesResponse, error) {
	nodes, err := c.ListNodes("FABRIC", page, limit)
	if err != nil {
		return nil, err
	}

	// Filter only orderer nodes
	var ordererNodes httptypes.ListNodesResponse
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

// ListBesuNodes lists all Besu nodes
func (c *Client) ListBesuNodes(page, limit int) (*httptypes.ListNodesResponse, error) {
	nodes, err := c.ListNodes("BESU", page, limit)
	if err != nil {
		return nil, err
	}

	// Filter only Besu nodes
	var besuNodes httptypes.ListNodesResponse
	for _, node := range nodes.Items {
		if node.NodeType == "BESU_FULLNODE" {
			besuNodes.Items = append(besuNodes.Items, node)
		}
	}
	besuNodes.Total = int64(len(besuNodes.Items))
	besuNodes.Page = nodes.Page
	besuNodes.PageCount = nodes.PageCount
	besuNodes.HasNextPage = nodes.HasNextPage

	return &besuNodes, nil
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
func (c *Client) UpdatePeerNode(id int64, req *types.FabricPeerConfig) (*httptypes.NodeResponse, error) {
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

	var node httptypes.NodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &node, nil
}

// UpdateOrdererNode updates a Fabric orderer node
func (c *Client) UpdateOrdererNode(id int64, req *types.FabricOrdererConfig) (*httptypes.NodeResponse, error) {
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

	var node httptypes.NodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &node, nil
}

// UpdateBesuNode updates a Besu node
func (c *Client) UpdateBesuNode(id int64, req *types.BesuNodeConfig) (*httptypes.NodeResponse, error) {
	body := map[string]interface{}{
		"blockchainPlatform": "BESU",
		"besuNode":           req,
	}

	resp, err := c.Put(fmt.Sprintf("/nodes/%d", id), body)
	if err != nil {
		return nil, fmt.Errorf("failed to update besu node: %w", err)
	}

	if err := CheckResponse(resp, stdhttp.StatusOK); err != nil {
		return nil, fmt.Errorf("failed to update besu node: %w", err)
	}

	var node httptypes.NodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &node, nil
}

// GetNode gets a node by ID
func (c *Client) GetNode(id int64) (*httptypes.NodeResponse, error) {
	resp, err := c.Get(fmt.Sprintf("/nodes/%d", id))
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	if err := CheckResponse(resp, stdhttp.StatusOK); err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	var node httptypes.NodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &node, nil
}
