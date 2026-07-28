package main

import (
	"fmt"

	"github.com/azzimoda/subscriberest/internal/handler"
	"github.com/azzimoda/subscriberest/internal/model"
	"github.com/azzimoda/subscriberest/internal/repository"
	"github.com/azzimoda/subscriberest/internal/router"
	"github.com/azzimoda/subscriberest/internal/service"
	"github.com/azzimoda/subscriberest/pkg/config"
	"github.com/azzimoda/subscriberest/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

// TODO: Swagger documentation.

func main() {
	config.Init()

	if viper.GetBool(config.KProductionMode) {
		gin.SetMode(gin.ReleaseMode)
	}

	logger.Init(viper.GetString(config.KLogLevel), viper.GetString(config.KLogDir))

	host := viper.GetString(config.KDBHost)
	port := viper.GetString(config.KDBPort)
	user := viper.GetString(config.KDBUser)
	password := viper.GetString(config.KDBPassword)
	dbName := viper.GetString(config.KDBName)
	sslMode := viper.GetString(config.KDBSSLMode)

	log.Debug().Str("host", host).Str("port", port).Str("user", user).Str("dbName", dbName).Str("sslMode", sslMode).
		Msg("Database configuration")

	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s", host, port, user, dbName, sslMode)
	log.Debug().Str("dsn", dsn).Msg("Connecting to database")

	if password != "" {
		dsn = fmt.Sprintf("%s password=%s", dsn, password)
		log.Trace().Msg("Using DB password")
	} else {
		log.Trace().Msg("No DB password")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: glogger.Default.LogMode(glogger.Info),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect database!")
	}

	if err := db.AutoMigrate(&model.Subscription{}); err != nil {
		log.Fatal().Err(err).Msg("Failed to auto-migrate!")
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get sql db!")
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	repository := repository.NewSubscriptionRepository(db)
	service := service.NewService(repository)
	handler := handler.NewHandler(service)
	engine := router.Init(handler)

	if err := engine.Run(":" + viper.GetString(config.KPort)); err != nil {
		log.Fatal().Err(err).Msg("")
	}
}
