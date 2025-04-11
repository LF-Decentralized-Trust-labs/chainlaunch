package types

import "time"

// Channel represents a Fabric channel
type Channel struct {
	Name      string    `json:"name"`
	BlockNum  int64     `json:"blockNum"`
	CreatedAt time.Time `json:"createdAt"`
}
