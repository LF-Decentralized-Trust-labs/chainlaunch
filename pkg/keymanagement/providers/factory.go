package providers

import (
	"fmt"

	"github.com/chainlaunch/chainlaunch/pkg/db"
	"github.com/chainlaunch/chainlaunch/pkg/keymanagement/providers/database"
)

type ProviderType string

const (
	ProviderTypeDatabase       ProviderType = "DATABASE"
	ProviderTypeHSM            ProviderType = "HSM"
	ProviderTypeHashicorpVault ProviderType = "HASHICORP_VAULT"
)

// ProviderFactory creates and manages key providers
type ProviderFactory struct {
	providers map[ProviderType]Provider
	queries   *db.Queries
}

func NewProviderFactory(queries *db.Queries) (*ProviderFactory, error) {
	factory := &ProviderFactory{
		providers: make(map[ProviderType]Provider),
		queries:   queries,
	}

	// Initialize database provider
	dbProvider, err := database.NewDatabaseProvider(queries)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database provider: %w", err)
	}
	factory.providers[ProviderTypeDatabase] = dbProvider

	return factory, nil
}

func (f *ProviderFactory) GetProvider(providerType ProviderType) (Provider, error) {
	provider, ok := f.providers[providerType]
	if !ok {
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
	return provider, nil
}
