package config

import (
	"github.com/spf13/viper"
)

const (
	KProductionMode = "production_mode"

	KLogLevel = "log_level"
	KLogDir   = "log_dir"

	KDBHost     = "db_host"
	KDBPort     = "db_port"
	KDBUser     = "db_user"
	KDBPassword = "db_password"
	KDBName     = "db_name"
	KDBSSLMode  = "db_sslmode"

	KPort = "port"
)

func Init() {
	viper.SetDefault(KProductionMode, false) // Set to true in production environment

	viper.SetDefault(KLogLevel, "debug")
	viper.SetDefault(KLogDir, "") // Empty value means no logging to file.

	viper.SetDefault(KPort, "8080")

	viper.AutomaticEnv()
}
