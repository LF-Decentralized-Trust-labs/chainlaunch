package besu

// NodeType represents the type of a Besu node
type NodeType string

const (
	NodeTypeValidator NodeType = "validator"
	NodeTypeMember    NodeType = "member"
)

// NodeStatus represents the status of a Besu node
type NodeStatus string

const (
	NodeStatusCreating NodeStatus = "creating"
	NodeStatusRunning  NodeStatus = "running"
	NodeStatusStopped  NodeStatus = "stopped"
	NodeStatusError    NodeStatus = "error"
)

// BesuNode represents a node in a Besu network
type BesuNode struct {
	ID             int64    `json:"id"`
	Type           NodeType `json:"type"`
	Address        string   `json:"address"`
	PublicKey      string   `json:"publicKey"`
	ValidatorIndex int      `json:"validatorIndex,omitempty"`
}

// GenesisParams represents the Besu genesis file configuration
type GenesisParams struct {
	Config     Config                       `json:"config"`
	Nonce      string                       `json:"nonce"`
	Timestamp  string                       `json:"timestamp"`
	GasLimit   string                       `json:"gasLimit"`
	Difficulty string                       `json:"difficulty"`
	MixHash    string                       `json:"mixHash"`
	Coinbase   string                       `json:"coinbase"`
	Alloc      map[string]map[string]string `json:"alloc"`
	ExtraData  string                       `json:"extraData"`
	Number     string                       `json:"number"`
	GasUsed    string                       `json:"gasUsed"`
	ParentHash string                       `json:"parentHash"`
}

// QBFTConfig represents the QBFT consensus configuration
type QBFTConfig struct {
	BlockPeriodSeconds    int   `json:"blockperiodseconds"`
	EpochLength           int   `json:"epochlength"`
	RequestTimeoutSeconds int   `json:"requesttimeoutseconds"`
	StartBlock            int64 `json:"startBlock"`
}

// Config represents the chain configuration
type Config struct {
	ChainID     int64      `json:"chainId"`
	BerlinBlock int        `json:"berlinBlock"`
	QBFT        QBFTConfig `json:"qbft"`
}

// NetworkConfig represents the configuration for a Besu network
type NetworkConfig struct {
	ChainID           int64   `json:"chainId"`
	BlockPeriod       int     `json:"blockPeriod"`
	EpochLength       int     `json:"epochLength"`
	RequestTimeout    int     `json:"requestTimeout"`
	Nonce             string  `json:"nonce"`
	Timestamp         string  `json:"timestamp"`
	GasLimit          string  `json:"gasLimit"`
	Difficulty        string  `json:"difficulty"`
	MixHash           string  `json:"mixHash"`
	Coinbase          string  `json:"coinbase"`
	InitialValidators []int64 `json:"initialValidators"`
}
