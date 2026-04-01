package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"web_backend/internal/app/config"
	"web_backend/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
	Config     *config.Config
}

func NewHandler(r *repository.Repository, cfg *config.Config) *Handler {
	return &Handler{
		Repository: r,
		Config:     cfg,
	}
}

// RegisterHandler godoc
// @title Heating API
// @version 1.0
// @description API для управления заявками на системную нагрузку
// @contact.name API Support
// @contact.url http://localhost:8080
// @contact.email support@example.com
// @license.name MIT
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.Use(CORSMiddleware())

	api := router.Group("/api")

	unauthorized := api.Group("/")
	unauthorized.POST("/users/register", h.CreateUser)
	unauthorized.POST("/users/login", h.SignIn)
	unauthorized.GET("/components", h.GetComponents)
	unauthorized.GET("/components/:id", h.GetComponent)

	optionalAuth := api.Group("/")
	optionalAuth.Use(h.WithOptionalAuthCheck())
	optionalAuth.GET("/heatings/cart", h.GetHeatingCart)

	authorized := api.Group("/")
	authorized.Use(h.ModeratorMiddleware(false))
	authorized.POST("/components", h.CreateComponent)
	authorized.GET("/heatings", h.GetAllHeatings)
	authorized.GET("/heatings/:id", h.GetHeating)
	authorized.PUT("/heatings/:id", h.EditHeating)
	authorized.PUT("/heatings/:id/form", h.FormHeating)
	authorized.DELETE("/heatings/:id", h.DeleteHeating)
	authorized.POST("/heating_components/add/:component_id", h.AddToHeating)
	authorized.DELETE("/heating_components/:component_id/:heating_id", h.DeleteFromHeating)
	authorized.PUT("/heating_components/:component_id/:heating_id", h.EditInHeating)
	authorized.POST("/users/logout", h.SignOut)

	moderator := api.Group("/")
	moderator.Use(h.ModeratorMiddleware(true))
	moderator.PUT("/heatings/:id/finish", h.FinishHeating)

	swaggerURL := ginSwagger.URL("/swagger/doc.json")
	router.Any("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, swaggerURL))
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.Static("/static", "./resources")
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	var errorMessage string
	switch {
	case errors.Is(err, repository.ErrNotFound):
		errorMessage = "Не найден"
	case errors.Is(err, repository.ErrAlreadyExists):
		errorMessage = "Уже существует"
	case errors.Is(err, repository.ErrNotAllowed):
		errorMessage = "Доступ запрещен"
	case errors.Is(err, repository.ErrNoDraft):
		errorMessage = "Черновик не найден"
	default:
		errorMessage = err.Error()
	}
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": errorMessage,
	})
}
