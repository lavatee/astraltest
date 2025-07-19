package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/lavatee/astraltest"
	"github.com/lavatee/astraltest/internal/endpoint"
	"github.com/lavatee/astraltest/internal/repository"
	"github.com/lavatee/astraltest/internal/service"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{PrettyPrint: true})
	if err := InitConfig(); err != nil {
		logrus.Fatalf("Failed to load config: %s", err.Error())
	}
	db, err := repository.NewPostgresDB(repository.PostgresConfig{
		Host:     viper.GetString("postgres.host"),
		Port:     viper.GetString("postgres.port"),
		User:     viper.GetString("postgres.user"),
		Password: viper.GetString("postgres.password"),
		DBName:   viper.GetString("postgres.dbname"),
		SSLMode:  viper.GetString("postgres.sslmode"),
	})
	if err != nil {
		logrus.Fatalf("Failed to connect Postgres DB: %s", err.Error())
	}
	defer db.Close()
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		logrus.Fatalf("Failed to create migrate driver: %s", err.Error())
	}

	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	migrationsPath := "file://" + filepath.Join(dir, "../schema")
	migrations, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	if err != nil {
		logrus.Fatalf("Failed to create migrate instance: %s", err.Error())
	}
	if err = migrations.Up(); err != nil && err != migrate.ErrNoChange {
		logrus.Fatalf("Migrations error: %s", err.Error())
	}

	repo := repository.NewRepository(db)
	intRedisDB, err := strconv.Atoi(viper.GetString("redis.db"))
	if err != nil {
		logrus.Fatal("Invalid value of redis.db")
	}
	cache, err := service.NewRedisClient(service.RedisConfig{
		Host:     viper.GetString("redis.host"),
		Port:     viper.GetString("redis.port"),
		Password: viper.GetString("redis.password"),
		DB:       intRedisDB,
	})
	if err != nil {
		logrus.Fatalf("Failed to connect Redis: %s", err.Error())
	}
	defer cache.Close()
	services := service.NewService(repo, viper.GetString("adminToken"), cache)
	endp := endpoint.NewEndpoint(services)
	server := &astraltest.Server{}
	go func() {
		if err := server.Run(viper.GetString("port"), endp.InitRoutes()); err != nil {
			logrus.Fatalf("Failed to run server: %s", err.Error())
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	if err := server.Shutdown(context.Background()); err != nil {
		logrus.Fatalf("server shutdown error: %s", err.Error())
	}
}

func InitConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
