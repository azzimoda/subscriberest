//go:generate swag init -d ../.. -g cmd/server/main.go -o ../../docs --parseDependency

// @title           SubscribeREST API
// @version         1.0
// @description     API для управления подписками пользователей
// @host            localhost:8080
// @BasePath        /api/v1

package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/azzimoda/subscriberest/internal/handler"
	_ "github.com/azzimoda/subscriberest/docs"
	"github.com/azzimoda/subscriberest/internal/repository"
	"github.com/azzimoda/subscriberest/internal/router"
	"github.com/azzimoda/subscriberest/internal/service"
	"github.com/azzimoda/subscriberest/pkg/config"
	"github.com/azzimoda/subscriberest/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

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

	migrateURL := (&url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   fmt.Sprintf("%s:%s", host, port),
		Path:   dbName,
		RawQuery: fmt.Sprintf("sslmode=%s", sslMode),
	}).String()
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get working directory")
	}
	m, err := migrate.New("file://"+filepath.Join(pwd, "migrations"), migrateURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create migrator")
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}
	sourceErr, dbErr := m.Close()
	if sourceErr != nil {
		log.Error().Err(sourceErr).Msg("Failed to close migration source")
	}
	if dbErr != nil {
		log.Error().Err(dbErr).Msg("Failed to close migration database")
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

	srv := &http.Server{
		Addr:    ":" + viper.GetString(config.KPort),
		Handler: engine,
	}

	go func() {
		log.Info().Str("port", viper.GetString(config.KPort)).Msg("Starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}
