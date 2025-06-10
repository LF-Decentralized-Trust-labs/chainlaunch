package projectrunner

// BoilerplateRunnerConfig defines how to run a project for a given boilerplate type.
type BoilerplateRunnerConfig struct {
	Command string
	Args    []string
	Image   string // Docker image to use for this boilerplate
}

var boilerplateRunners = map[string]BoilerplateRunnerConfig{
	"server-ts": {
		Command: "bun",
		Args:    []string{"--watch", "run", "index.ts"},
		Image:   "oven/bun:latest",
	},
	"chaincode-fabric-ts": {
		Args:  []string{"npm", "run", "start:dev"},
		Image: "chaincode-ts:1.0",
	},
}

// GetBoilerplateRunner returns the command, args, and image for a given boilerplate type.
func GetBoilerplateRunner(boilerplate string) (string, []string, string, bool) {
	runner, ok := boilerplateRunners[boilerplate]
	return runner.Command, runner.Args, runner.Image, ok
}
