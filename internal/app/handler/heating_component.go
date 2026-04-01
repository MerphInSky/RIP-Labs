package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"web_backend/internal/app/repository"
	"web_backend/internal/app/serializer"
)

// AddToHeating godoc
// @Summary Добавить компонент в заявку
// @Description Добавляет компонент в заявку-черновик пользователя
// @Tags heating_components
// @Produce json
// @Param component_id path int true "ID компонента"
// @Success 200 {object} serializer.HeatingJSON "Заявка с добавленным компонентом"
// @Success 201 {object} serializer.HeatingJSON "Создана новая заявка"
// @Failure 400 {object} map[string]string "Неверный запрос"
// @Failure 404 {object} map[string]string "Компонент не найден"
// @Failure 409 {object} map[string]string "Компонент уже в заявке"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /api/heating_components/add/{component_id} [post]
func (h *Handler) AddToHeating(ctx *gin.Context) {
	componentIDStr := ctx.Param("component_id")
	componentID, err := strconv.Atoi(componentIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	creatorID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}
	load, created, err := h.Repository.GetHeatingDraft(creatorID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	err = h.Repository.AddComponent(uint(componentID), creatorID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if errors.Is(err, repository.ErrAlreadyExists) {
			h.errorHandler(ctx, http.StatusConflict, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}
	creatorLogin, moderatorLogin, _ := h.Repository.GetModeratorAndCreatorLogin(load)
	completedCount, _ := h.Repository.GetCompletedItemCount(load.HeatingID)
	status := http.StatusOK
	if created {
		ctx.Header("Location", fmt.Sprintf("/api/heatings/%d", load.HeatingID))
		status = http.StatusCreated
	}
	ctx.JSON(status, serializer.HeatingToJSON(load, creatorLogin, moderatorLogin, completedCount))
}

// DeleteFromHeating godoc
// @Summary Удалить компонент из заявки
// @Description Удаляет связь компонента и заявки (только черновик)
// @Tags heating_components
// @Produce json
// @Param component_id path int true "ID компонента"
// @Param heating_id path int true "ID заявки"
// @Success 200 {object} serializer.HeatingJSON "Обновленная заявка"
// @Failure 400 {object} map[string]string "Неверные ID"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Не найдено"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /api/heating_components/{component_id}/{heating_id} [delete]
func (h *Handler) DeleteFromHeating(ctx *gin.Context) {
	componentID, err := strconv.Atoi(ctx.Param("component_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	heatingID, err := strconv.Atoi(ctx.Param("heating_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	load, err := h.Repository.DeleteComponentFromHeating(heatingID, componentID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if errors.Is(err, repository.ErrNotAllowed) {
			h.errorHandler(ctx, http.StatusForbidden, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}
	creatorLogin, moderatorLogin, _ := h.Repository.GetModeratorAndCreatorLogin(load)
	completedCount, _ := h.Repository.GetCompletedItemCount(load.HeatingID)
	ctx.JSON(http.StatusOK, serializer.HeatingToJSON(load, creatorLogin, moderatorLogin, completedCount))
}

// EditInHeating godoc
// @Summary Изменить данные компонента в заявке
// @Description Обновляет параметры компонента в конкретной заявке (только черновик)
// @Tags heating_components
// @Accept json
// @Produce json
// @Param component_id path int true "ID компонента"
// @Param heating_id path int true "ID заявки"
// @Param data body serializer.HeatingComponentJSON true "Новые данные"
// @Success 200 {object} serializer.HeatingComponentJSON "Обновленные данные"
// @Failure 400 {object} map[string]string "Неверные данные"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Не найдено"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /api/heating_components/{component_id}/{heating_id} [put]
func (h *Handler) EditInHeating(ctx *gin.Context) {
	componentID, err := strconv.Atoi(ctx.Param("component_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	heatingID, err := strconv.Atoi(ctx.Param("heating_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	var j serializer.HeatingComponentJSON
	if err := ctx.BindJSON(&j); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	item, err := h.Repository.EditComponentInHeating(heatingID, componentID, j)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.errorHandler(ctx, http.StatusNotFound, err)
		} else if errors.Is(err, repository.ErrNotAllowed) {
			h.errorHandler(ctx, http.StatusForbidden, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}
	ctx.JSON(http.StatusOK, serializer.HeatingComponentToJSON(item))
}
