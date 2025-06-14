package projectrunner

import (
	"fmt"

	"github.com/chainlaunch/chainlaunch/pkg/scai/boilerplates"
)

// BoilerplateRunnerConfig defines how to run a project for a given boilerplate type.
type BoilerplateRunnerConfig struct {
	Command string
	Args    []string
	Image   string // Docker image to use for this boilerplate
}

var boilerplateRunners = map[string]BoilerplateRunnerConfig{
	"chaincode-fabric-ts": {
		Args:  []string{"npm", "run", "start:dev"},
		Image: "chaincode-ts:1.0",
	},
}

// GetBoilerplateRunner returns the command, args, and image for a given boilerplate type
func GetBoilerplateRunner(boilerplateService *boilerplates.BoilerplateService, boilerplateType string) (string, []string, string, error) {
	config, err := boilerplateService.GetBoilerplateConfig(boilerplateType)
	if err != nil {
		return "", nil, "", fmt.Errorf("unknown boilerplate type: %s", boilerplateType)
	}

	return config.Command, config.Args, config.Image, nil
}
