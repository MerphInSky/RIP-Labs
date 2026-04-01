package main

import (
	"web_backend/internal/app/config"
	"web_backend/internal/app/dsn"
	"web_backend/internal/app/handler"
	"web_backend/internal/app/repository"
	"web_backend/internal/pkg"

	_ "web_backend/docs"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// @title Heating API
// @version 1.0
// @description API для расчёта нагрева компонентов
// @host localhost:8080
// @BasePath /
// @SecurityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @Security ApiKeyAuth

func main() {
	router := gin.Default()

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	logrus.Info("DSN: ", postgresString)

	rep, err := repository.New(postgresString)
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}

	hand := handler.NewHandler(rep, conf)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
