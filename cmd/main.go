package main

import (
	"log"
	"os"

	goproj "github.com/gfifgfifofich/GoProj"
	"github.com/gfifgfifofich/GoProj/pkg/handler"
	"github.com/gfifgfifofich/GoProj/pkg/repository"
	"github.com/gfifgfifofich/GoProj/pkg/service"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

func main() {
	if err := initConfig(); err != nil {
		log.Fatalf("error initializing config %s", err.Error())
	}

	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading environment file: %s", err.Error())
	}

	db, err := repository.NewDatabase(repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
	})
	if err != nil {
		log.Fatalf("error connecting to database: %s", err.Error())
	}

	repos := repository.NewRepository(db)
	services := service.NewService(repos)
	handlers := handler.NewHandler(services)
	server := new(goproj.Server)
	if err := server.Run(viper.GetString("port"), handlers.InitRoutes()); err != nil {
		log.Fatalf("error while running server: %s", err.Error())
	}
}

func initConfig() error {
	viper.AddConfigPath("Config")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
