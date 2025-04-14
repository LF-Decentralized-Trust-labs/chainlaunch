package service

import (
	"time"
)

type ImportNetworkParams struct {
	NetworkType string
	GenesisFile []byte
	Name        string
	Description string
	ChainID     *int64
}

type ImportNetworkResult struct {
	NetworkID string
	Message   string
}

// JoinPeerRequest represents a request to join a peer to a network
type JoinPeerRequest struct {
	PeerID  int64  `json:"peer_id" validate:"required"`
	OrgName string `json:"org_name" validate:"required"`
}

// JoinPeerResponse represents a response from joining a peer to a network
type JoinPeerResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	NetworkID int64  `json:"network_id"`
	PeerID    int64  `json:"peer_id"`
}

// JoinOrdererRequest represents a request to join an orderer to a network
type JoinOrdererRequest struct {
	OrdererID int64  `json:"orderer_id" validate:"required"`
	OrgName   string `json:"org_name" validate:"required"`
}

// JoinOrdererResponse represents a response from joining an orderer to a network
type JoinOrdererResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	NetworkID int64  `json:"network_id"`
	OrdererID int64  `json:"orderer_id"`
}

// RemovePeerRequest represents a request to remove a peer from a network
type RemovePeerRequest struct {
	PeerID int64 `json:"peer_id" validate:"required"`
}

// RemovePeerResponse represents a response from removing a peer from a network
type RemovePeerResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	NetworkID int64  `json:"network_id"`
	PeerID    int64  `json:"peer_id"`
}

// RemoveOrdererRequest represents a request to remove an orderer from a network
type RemoveOrdererRequest struct {
	OrdererID int64 `json:"orderer_id" validate:"required"`
}

// RemoveOrdererResponse represents a response from removing an orderer from a network
type RemoveOrdererResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	NetworkID int64  `json:"network_id"`
	OrdererID int64  `json:"orderer_id"`
}

// Block represents a block in the blockchain
type Block struct {
	Number       uint64    `json:"number"`
	Hash         string    `json:"hash"`
	PreviousHash string    `json:"previous_hash"`
	Timestamp    time.Time `json:"timestamp"`
	TxCount      int       `json:"tx_count"`
	Data         []byte    `json:"data"`
}

// Transaction represents a transaction in a block
type Transaction struct {
	TxID        string    `json:"tx_id"`
	BlockNumber uint64    `json:"block_number"`
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"`
	Creator     string    `json:"creator"`
	Payload     []byte    `json:"payload"`
}

type ChainInfo struct {
	Height            uint64
	CurrentBlockHash  string
	PreviousBlockHash string
}
