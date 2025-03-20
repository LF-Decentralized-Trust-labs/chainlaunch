package config

import "embed"

type ConfigCMD struct {
	Views        embed.FS
	MigrationsFS embed.FS
}
