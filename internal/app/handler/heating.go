package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"web_backend/internal/app/repository"
	"web_backend/internal/app/serializer"
)

// GetHeatingCart godoc
// @Summary Получить корзину заявки
// @Description Возвращает информацию о текущей заявке-черновике пользователя
// @Tags heatings
// @Produce json
// @Success 200 {object} map[string]interface{} "Данные корзины"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/heatings/cart [get]
func (h *Handler) GetHeatingCart(ctx *gin.Context) {
	creatorID, err := getUserID(ctx)
	if err != nil || creatorID == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"has_draft":        false,
			"components_count": 0,
		})
		return
	}
	count := h.Repository.GetHeatingComponentCount(creatorID)
	if count == 0 {
		load, err := h.Repository.CheckCurrentDraft(creatorID)
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"has_draft":        false,
				"components_count": 0,
			})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"id":               load.HeatingID,
			"has_draft":        true,
			"components_count": 0,
		})
		return
	}
	loadID := h.Repository.GetActiveHeatingID(creatorID)
	ctx.JSON(http.StatusOK, gin.H{
		"id":               loadID,
		"has_draft":        true,
		"components_count": count,
	})
}

// GetAllHeatings godoc
// @Summary Получить список заявок
// @Description Возвращает заявки с фильтрацией по датам и статусу. Создатель видит только свои, модератор — все.
// @Tags heatings
// @Produce json
// @Param from-date query string false "Начальная дата (YYYY-MM-DD)"
// @Param to-date query string false "Конечная дата (YYYY-MM-DD)"
// @Param status query string false "Статус заявки"
// @Success 200 {array} serializer.HeatingJSON "Список заявок"
// @Failure 400 {object} map[string]string "Неверный формат даты"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /api/heatings [get]
func (h *Handler) GetAllHeatings(ctx *gin.Context) {
	fromDate := ctx.Query("from-date")
	var from, to time.Time
	if fromDate != "" {
		t, err := time.Parse("2006-01-02", fromDate)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, err)
			return
		}
		from = t
	}
	toDate := ctx.Query("to-date")
	if toDate != "" {
		t, err := time.Parse("2006-01-02", toDate)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, err)
			return
		}
		to = t
	}
	status := ctx.Query("status")

	userID, _ := getUserID(ctx)
	creatorID := uint(0)
	if isModVal, ok := ctx.Get("is_moderator"); ok {
		if isMod, _ := isModVal.(bool); !isMod {
			creatorID = userID
		}
	} else {
		creatorID = userID
	}

	loads, err := h.Repository.GetAllHeatings(from, to, status, creatorID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	resp := make([]serializer.HeatingJSON, 0, len(loads))
	for _, load := range loads {
		creatorLogin, moderatorLogin, _ := h.Repository.GetModeratorAndCreatorLogin(load)
		completedCount, _ := h.Repository.GetCompletedItemCount(load.HeatingID)
		resp = append(resp, serializer.HeatingToJSON(load, creatorLogin, moderatorLogin, completedCount))
	}
	ctx.JSON(http.StatusOK, resp)
}

// GetHeating godoc
// @Summary Получить заявку по ID
// @Description Возвращает полную информацию о заявке с компонентами
// @Tags heatings
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} map[string]interface{} "Данные заявки с компонентами"
// @Failure 400 {object} map[string]string "Неверный ID"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Заявка не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /api/heatings/{id} [get]
func (h *Handler) GetHeating(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	load, err := h.Repository.GetSingleHeating(id)
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

	userID, _ := getUserID(ctx)
	if isModVal, ok := ctx.Get("is_moderator"); ok {
		if isMod, _ := isModVal.(bool); !isMod && load.CreatorID != userID {
			h.errorHandler(ctx, http.StatusForbidden, repository.ErrNotAllowed)
			return
		}
	} else if load.CreatorID != userID {
		h.errorHandler(ctx, http.StatusForbidden, repository.ErrNotAllowed)
		return
	}

	items, err := h.Repository.GetHeatingItems(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	creatorLogin, moderatorLogin, _ := h.Repository.GetModeratorAndCreatorLogin(load)
	completedCount, _ := h.Repository.GetCompletedItemCount(load.HeatingID)
	itemsResp := make([]serializer.HeatingComponentDetailJSON, 0, len(items))
	for _, item := range items {
		itemsResp = append(itemsResp, serializer.HeatingComponentDetailToJSON(item))
	}
	ctx.JSON(http.StatusOK, gin.H{
		"heating": serializer.HeatingToJSON(load, creatorLogin, moderatorLogin, completedCount),
		"components":  itemsResp,
	})
}

// EditHeating godoc
// @Summary Изменить заявку
// @Description Обновляет данные заявки (только черновик)
// @Tags heatings
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param body body serializer.HeatingJSON true "Новые данные заявки"
// @Success 200 {object} serializer.HeatingJSON "Обновленная заявка"
// @Failure 400 {object} map[string]string "Неверные данные"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Заявка не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /api/heatings/{id} [put]
func (h *Handler) EditHeating(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	var j serializer.HeatingJSON
	if err := ctx.BindJSON(&j); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	load, err := h.Repository.EditHeating(id, j)
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

// FormHeating godoc
// @Summary Сформировать заявку
// @Description Переводит заявку из черновика в статус "formed"
// @Tags heatings
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} serializer.HeatingJSON "Сформированная заявка"
// @Failure 400 {object} map[string]string "Неверный запрос"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Заявка не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /api/heatings/{id}/form [put]
func (h *Handler) FormHeating(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	creatorID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}
	load, err := h.Repository.FormHeating(id, creatorID)
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

// FinishHeating godoc
// @Summary Завершить заявку (только модератор)
// @Description Изменяет статус заявки на completed/rejected
// @Tags heatings
// @Accept json
// @Produce json
// @Param id path int true "ID заявки"
// @Param status body serializer.StatusJSON true "Новый статус (completed/rejected)"
// @Success 200 {object} serializer.HeatingJSON "Результат модерации"
// @Failure 400 {object} map[string]string "Неверный запрос"
// @Failure 403 {object} map[string]string "Доступ запрещен (не модератор)"
// @Failure 404 {object} map[string]string "Заявка не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /api/heatings/{id}/finish [put]
func (h *Handler) FinishHeating(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	moderatorID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}
	var statusJSON serializer.StatusJSON
	if err := ctx.BindJSON(&statusJSON); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	load, err := h.Repository.FinishHeating(id, statusJSON.Status, moderatorID)
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

// DeleteHeating godoc
// @Summary Удалить заявку
// @Description Логическое удаление заявки (только черновик, только создатель)
// @Tags heatings
// @Produce json
// @Param id path int true "ID заявки"
// @Success 200 {object} map[string]string "Заявка удалена"
// @Failure 400 {object} map[string]string "Неверный запрос"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Заявка не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /api/heatings/{id} [delete]
func (h *Handler) DeleteHeating(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	creatorID, err := getUserID(ctx)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}
	_, err = h.Repository.DeleteHeating(id, creatorID)
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
	ctx.JSON(http.StatusOK, gin.H{"message": "Заявка удалена"})
}
