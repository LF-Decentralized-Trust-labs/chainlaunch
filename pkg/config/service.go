package config

// ConfigService handles configuration paths and directories
type ConfigService struct {
	dataPath string
}

// NewConfigService creates a new ConfigService instance
func NewConfigService(dataPath string) *ConfigService {
	return &ConfigService{
		dataPath: dataPath,
	}
}

func (s *ConfigService) GetDataPath() string {
	return s.dataPath
}
