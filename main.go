/*
Copyright © 2025 ChainLaunch <dviejo@chainlaunch.dev>
*/
package main

import (
	"embed"

	"github.com/chainlaunch/chainlaunch/cmd"
	"github.com/chainlaunch/chainlaunch/config"
	"github.com/sirupsen/logrus"
)

//go:embed web/dist/*
var views embed.FS

//go:embed pkg/db/*
var migrationsFS embed.FS

func main() {
	configCMD := config.ConfigCMD{
		Views:        views,
		MigrationsFS: migrationsFS,
	}
	logrus.SetLevel(logrus.DebugLevel)
	err := cmd.NewRootCmd(configCMD).Execute()
	if err != nil {
		logrus.Fatalf("Error executing command: %v", err)
	}
}
