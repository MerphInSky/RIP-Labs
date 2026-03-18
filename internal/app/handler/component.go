package handler

import (
	"net/http"
	"strconv"
	"web_backend/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetComponents(ctx *gin.Context) {
	var components []ds.Component
	var err error

	searchQuery := ctx.Query("query")
	if searchQuery == "" {
		components, err = h.Repository.GetComponents()
	} else {
		components, err = h.Repository.GetComponentsByTitle(searchQuery)
	}
	if err != nil {
		logrus.Error(err)
	}

	creatorID := uint(1)
	loadCount := h.Repository.GetHeatingComponentCount(creatorID)
	activeLoadID := h.Repository.GetActiveHeatingID(creatorID)

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"components":      components,
		"query":           searchQuery,
		"heating_count": loadCount,
		"heating_id":   activeLoadID,
		"minioUrl":         h.Config.MinioURL,
	})
}

func (h *Handler) GetComponent(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	component, err := h.Repository.GetComponent(id)
	if err != nil {
		logrus.Error(err)
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}

	ctx.HTML(http.StatusOK, "component.html", gin.H{
		"component": component,
		"minioUrl": h.Config.MinioURL,
	})
}
