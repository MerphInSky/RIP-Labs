package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetHeating(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	creatorID := uint(1)
	isDraft, err := h.Repository.IsDraftHeating(id, creatorID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	if !isDraft {
		ctx.Redirect(http.StatusSeeOther, "/")
		return
	}

	items, load, err := h.Repository.GetHeating(id, creatorID)
	if err != nil {
		logrus.Error(err)
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.HTML(http.StatusOK, "heating.html", gin.H{
		"items":          items,
		"load":           load,
		"heating_id": id,
		"minioUrl":       h.Config.MinioURL,
	})
}

func (h *Handler) AddToHeating(ctx *gin.Context) {
	componentIDStr := ctx.PostForm("component_id")
	componentID, err := strconv.Atoi(componentIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	creatorID := uint(1)

	err = h.Repository.AddComponent(uint(componentID), creatorID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusSeeOther, ctx.Request.Referer())
}

func (h *Handler) DeleteHeating(ctx *gin.Context) {
	loadIDStr := ctx.PostForm("heating_id")
	loadID, err := strconv.Atoi(loadIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.DeleteHeating(uint(loadID))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/")
}
