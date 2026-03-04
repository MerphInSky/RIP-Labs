package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/", h.GetComponents)
	router.GET("/components/:id", h.GetComponent)
	router.GET("/heatings/:id", h.GetHeating)
	router.POST("/heatings/add", h.AddToHeating)
	router.POST("/heatings/delete", h.DeleteHeating)
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./resources")
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
