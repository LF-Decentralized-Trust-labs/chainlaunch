package http

import "encoding/json"

// BesuNetworkRequest represents the request to create a new Besu network
// @Description Request body for creating a new Besu network
type CreateBesuNetworkRequest struct {
	// @Description Name of the network
	Name string `json:"name" validate:"required"`
	// @Description Optional description of the network
	Description string `json:"description"`
	// @Description Network configuration
	Config struct {
		// @Description Consensus algorithm (e.g. "qbft")
		// @Required
		Consensus string `json:"consensus" validate:"required"`
		// @Description Chain ID for the network
		// @Default 1337
		// @Required
		ChainID int64 `json:"chainId" validate:"required" example:"1337"`
		// @Description Block period in seconds
		// @Default 5
		// @Required
		BlockPeriod int `json:"blockPeriod" validate:"required" example:"5"`
		// @Description Epoch length in blocks
		// @Default 30000
		// @Required
		EpochLength int `json:"epochLength" validate:"required" example:"30000"`
		// @Description Request timeout in seconds
		// @Required
		RequestTimeout int `json:"requestTimeout" validate:"required"`
		// @Description List of initial validator key IDs
		// @Required
		// @MinItems 1
		InitialValidatorKeyIds []int64 `json:"initialValidatorsKeyIds" validate:"required,min=1"`
		// @Description Optional nonce value
		Nonce string `json:"nonce,omitempty"`
		// @Description Optional timestamp value
		Timestamp string `json:"timestamp,omitempty"`
		// @Description Optional gas limit value
		GasLimit string `json:"gasLimit,omitempty"`
		// @Description Optional difficulty value
		Difficulty string `json:"difficulty,omitempty"`
		// @Description Optional mix hash value
		MixHash string `json:"mixHash,omitempty"`
		// @Description Optional coinbase address
		Coinbase string `json:"coinbase,omitempty"`
	} `json:"config" validate:"required"`
}

// BesuNetworkResponse represents a Besu network in responses
type BesuNetworkResponse struct {
	ID            int64           `json:"id"`
	Platform      string          `json:"platform"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	Status        string          `json:"status"`
	ChainID       int64           `json:"chainId"`
	CreatedAt     string          `json:"createdAt"`
	UpdatedAt     string          `json:"updatedAt,omitempty"`
	Config        json.RawMessage `json:"config,omitempty"`
	GenesisConfig string          `json:"genesisConfig,omitempty"`
}

// ListBesuNetworksResponse represents the response for listing Besu networks
type ListBesuNetworksResponse struct {
	Networks []BesuNetworkResponse `json:"networks"`
	Total    int64                 `json:"total"`
}
