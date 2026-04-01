package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"web_backend/internal/app/ds"
	"web_backend/internal/app/repository"
	"web_backend/internal/app/serializer"
)

// GetComponents godoc
// @Summary Получить список компонентов
// @Description Возвращает все компоненты или фильтрует по названию
// @Tags components
// @Produce json
// @Param Title query string false "Название компонента для поиска"
// @Success 200 {array} serializer.ComponentJSON "Список компонентов"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/components [get]
func (h *Handler) GetComponents(ctx *gin.Context) {
	var components []ds.Component
	var err error
	searchQuery := ctx.Query("Title")
	if searchQuery == "" {
		components, err = h.Repository.GetComponents()
	} else {
		components, err = h.Repository.GetComponentsByTitle(searchQuery)
	}
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	resp := make([]serializer.ComponentJSON, 0, len(components))
	for _, s := range components {
		resp = append(resp, serializer.ComponentToJSON(s))
	}
	ctx.JSON(http.StatusOK, resp)
}

// GetComponent godoc
// @Summary Получить компонент по ID
// @Description Возвращает информацию о компоненте по идентификатору
// @Tags components
// @Produce json
// @Param id path int true "ID компонента"
// @Success 200 {object} serializer.ComponentJSON "Данные компонента"
// @Failure 400 {object} map[string]string "Неверный ID"
// @Failure 404 {object} map[string]string "Компонент не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/components/{id} [get]
func (h *Handler) GetComponent(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	component, err := h.Repository.GetComponent(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}
	ctx.JSON(http.StatusOK, serializer.ComponentToJSON(*component))
}

// CreateComponent godoc
// @Summary Создать компонент
// @Description Создает новый компонент (с возможной загрузкой изображения/видео)
// @Tags components
// @Accept json
// @Produce json
// @Param component body serializer.ComponentJSON true "Данные нового компонента"
// @Success 201 {object} serializer.ComponentJSON "Созданный компонент"
// @Failure 400 {object} map[string]string "Неверные данные"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /api/components [post]
func (h *Handler) CreateComponent(ctx *gin.Context) {
	contentType := ctx.GetHeader("Content-Type")
	var j serializer.ComponentJSON
	if strings.HasPrefix(contentType, "application/json") {
		if err := ctx.BindJSON(&j); err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, err)
			return
		}
	} else {
		title := ctx.PostForm("title")
		desc := ctx.PostForm("description")
		if title == "" || desc == "" {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("title and description are required"))
			return
		}
		res := 1.0
		if v := ctx.PostForm("thermal_resistance"); v != "" {
			fmt.Sscanf(v, "%f", &res)
		}
		j = serializer.ComponentJSON{
			Title:                  title,
			Description:            desc,
			ThermalResistance:     res,
		}
	}

	component, err := h.Repository.CreateComponent(j)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if imageFile, err := ctx.FormFile("image"); err == nil {
		s, err := h.Repository.AddPhoto(ctx, int(component.ComponentID), imageFile)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
		component = *s
	}
	if videoFile, err := ctx.FormFile("video"); err == nil {
		s, err := h.Repository.AddVideo(ctx, int(component.ComponentID), videoFile)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
		component = *s
	}

	ctx.Header("Location", fmt.Sprintf("/api/components/%d", component.ComponentID))
	ctx.JSON(http.StatusCreated, serializer.ComponentToJSON(component))
}
