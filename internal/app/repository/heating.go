package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"web_backend/internal/app/ds"
	"web_backend/internal/app/serializer"
)

func CalculateHeat(ambientTemperature, thermalResistance, powerDissipation float64) float64 {
	if thermalResistance <= 0 || powerDissipation < 0 {
		return ambientTemperature
	}
	return ambientTemperature + powerDissipation*thermalResistance
}


func (r *Repository) GetHeatingComponentCount(creatorID uint) int64 {
	var loadID uint
	err := r.db.Model(&ds.Heating{}).
		Where("creator_id = ? AND status = ?", creatorID, "draft").
		Select("heating_id").First(&loadID).Error
	if err != nil {
		return 0
	}

	var count int64
	err = r.db.Model(&ds.HeatingComponent{}).
		Where("heating_id = ?", loadID).Count(&count).Error
	if err != nil {
		logrus.Error("error counting heating_components:", err)
	}
	return count
}

func (r *Repository) GetActiveHeatingID(creatorID uint) uint {
	var loadID uint
	err := r.db.Model(&ds.Heating{}).
		Where("creator_id = ? AND status = ?", creatorID, "draft").
		Select("heating_id").First(&loadID).Error
	if err != nil {
		return 0
	}
	return loadID
}


func (r *Repository) CheckCurrentDraft(creatorID uint) (ds.Heating, error) {
	if creatorID == 0 {
		return ds.Heating{}, ErrNotAllowed
	}
	var load ds.Heating
	res := r.db.Where("creator_id = ? AND status = ?", creatorID, "draft").Limit(1).Find(&load)
	if res.Error != nil {
		return ds.Heating{}, res.Error
	}
	if res.RowsAffected == 0 {
		return ds.Heating{}, ErrNoDraft
	}
	return load, nil
}

func (r *Repository) GetHeatingDraft(creatorID uint) (ds.Heating, bool, error) {
	load, err := r.CheckCurrentDraft(creatorID)
	if errors.Is(err, ErrNoDraft) {
		load = ds.Heating{
			Status:    "draft",
			CreatedAt: time.Now(),
			CreatorID: creatorID,
		}
		if err := r.db.Create(&load).Error; err != nil {
			return ds.Heating{}, false, err
		}
		return load, true, nil
	}
	if err != nil {
		return ds.Heating{}, false, err
	}
	return load, false, nil
}

func (r *Repository) GetModeratorAndCreatorLogin(load ds.Heating) (string, string, error) {
	var creator ds.Users
	if err := r.db.Where("user_id = ?", load.CreatorID).First(&creator).Error; err != nil {
		return "", "", err
	}
	var moderatorLogin string
	if load.ModeratorID != nil && *load.ModeratorID != 0 {
		var moderator ds.Users
		if err := r.db.Where("user_id = ?", *load.ModeratorID).First(&moderator).Error; err != nil {
			return "", "", err
		}
		moderatorLogin = moderator.Login
	}
	return creator.Login, moderatorLogin, nil
}

func (r *Repository) GetCompletedItemCount(loadID uint) (int, error) {
	var count int64
	err := r.db.Model(&ds.HeatingComponent{}).
		Where("heating_id = ? AND heat IS NOT NULL", loadID).
		Count(&count).Error
	return int(count), err
}

func (r *Repository) GetAllHeatings(from, to time.Time, status string, creatorID uint) ([]ds.Heating, error) {
	var loads []ds.Heating
	sub := r.db.Where("status != ? AND status != ?", "deleted", "draft")
	if !from.IsZero() {
		sub = sub.Where("forming_date >= ?", from)
	}
	if !to.IsZero() {
		sub = sub.Where("forming_date < ?", to.Add(time.Hour*24))
	}
	if status != "" {
		sub = sub.Where("status = ?", status)
	}
	if creatorID != 0 {
		sub = sub.Where("creator_id = ?", creatorID)
	}
	err := sub.Order("heating_id").Find(&loads).Error
	return loads, err
}

func (r *Repository) GetSingleHeating(id int) (ds.Heating, error) {
	if id < 0 {
		return ds.Heating{}, errors.New("неверное id")
	}
	var load ds.Heating
	err := r.db.Where("heating_id = ?", id).First(&load).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.Heating{}, fmt.Errorf("%w: заявка с id %d", ErrNotFound, id)
		}
		return ds.Heating{}, err
	}
	if load.Status == "deleted" {
		return ds.Heating{}, fmt.Errorf("%w: заявка удалена", ErrNotAllowed)
	}
	return load, nil
}

func (r *Repository) GetHeatingItems(loadID int) ([]ds.HeatingComponent, error) {
	var items []ds.HeatingComponent
	err := r.db.Where("heating_id = ?", loadID).
		Preload("Component").
		Find(&items).Error
	return items, err
}

func (r *Repository) AddComponent(componentID uint, creatorID uint) error {
	var heating ds.Heating

	err := r.db.Where("creator_id = ? AND status = ?", creatorID, "draft").
		First(&heating).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		heating = ds.Heating{
			Status:    "draft",
			CreatedAt: time.Now(),
			CreatorID: creatorID,
		}
		if err := r.db.Create(&heating).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	var component ds.Component
	if err := r.db.Where("component_id = ? AND is_deleted = ?", componentID, false).First(&component).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: компонент с id %d", ErrNotFound, componentID)
		}
		return err
	}

	var count int64
	r.db.Model(&ds.HeatingComponent{}).
		Where("heating_id = ? AND component_id = ?", heating.HeatingID, componentID).
		Count(&count)

	if count > 0 {
		return fmt.Errorf("%w: компонент %d уже в заявке %d", ErrAlreadyExists, componentID, heating.HeatingID)
	}

	heat := CalculateHeat(
		25,component.ThermalResistance, 5,
	)

	item := ds.HeatingComponent{
		HeatingID: heating.HeatingID,
		ComponentID:   componentID,
		PowerDissipation:   1000,
		Heat: &heat,
	}
	return r.db.Create(&item).Error
}

func (r *Repository) DeleteComponentFromHeating(heatingID, componentID int) (ds.Heating, error) {
	var load ds.Heating
	err := r.db.Where("heating_id = ?", heatingID).First(&load).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.Heating{}, fmt.Errorf("%w: заявка с id %d", ErrNotFound, heatingID)
		}
		return ds.Heating{}, err
	}
	if load.Status != "draft" {
		return ds.Heating{}, fmt.Errorf("%w: можно удалять только из черновика", ErrNotAllowed)
	}
	err = r.db.Where("heating_id = ? AND component_id = ?", heatingID, componentID).
		Delete(&ds.HeatingComponent{}).Error
	if err != nil {
		return ds.Heating{}, err
	}
	return load, nil
}

func (r *Repository) EditComponentInHeating(heatingID, componentID int, j serializer.HeatingComponentJSON) (ds.HeatingComponent, error) {
	var item ds.HeatingComponent
	err := r.db.Where("heating_id = ? AND component_id = ?", heatingID, componentID).
		First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.HeatingComponent{}, fmt.Errorf("%w: компонент в заявке", ErrNotFound)
		}
		return ds.HeatingComponent{}, err
	}

	var heating ds.Heating
	if err := r.db.Where("heating_id = ?", heatingID).First(&heating).Error; err != nil {
		return ds.HeatingComponent{}, err
	}
	if heating.Status != "draft" {
		return ds.HeatingComponent{}, fmt.Errorf("%w: можно редактировать только черновик", ErrNotAllowed)
	}

	updates := map[string]interface{}{
		"power_dissipation": j.PowerDissipation,
	}

	var component ds.Component
	if err := r.db.First(&component, componentID).Error; err == nil {
		heat := CalculateHeat(
			component.ThermalResistance,
			heating.AmbientTemperature,
			j.PowerDissipation,
		)
		updates["heat"] = heat
	}

	err = r.db.Model(&item).Updates(updates).Error
	if err != nil {
		return ds.HeatingComponent{}, err
	}
	r.db.Where("heating_id = ? AND component_id = ?", heatingID, componentID).
		Preload("Component").First(&item)
	return item, nil
}

func (r *Repository) EditHeating(id int, j serializer.HeatingJSON) (ds.Heating, error) {
	var load ds.Heating
	err := r.db.Where("heating_id = ? AND status != ?", id, "deleted").First(&load).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.Heating{}, fmt.Errorf("%w: заявка с id %d", ErrNotFound, id)
		}
		return ds.Heating{}, err
	}
	if load.Status != "draft" {
		return ds.Heating{}, fmt.Errorf("%w: можно редактировать только черновик", ErrNotAllowed)
	}
	updates := serializer.HeatingFromJSON(j)
	err = r.db.Model(&load).Updates(updates).Error
	if err != nil {
		return ds.Heating{}, err
	}
	r.db.Where("heating_id = ?", id).First(&load)
	return load, nil
}

func (r *Repository) FormHeating(id int, creatorID uint) (ds.Heating, error) {
	heating, err := r.GetSingleHeating(id)
	if err != nil {
		return ds.Heating{}, err
	}
	if heating.Status != "draft" {
		return ds.Heating{}, fmt.Errorf("%w: только черновик можно сформировать", ErrNotAllowed)
	}
	if heating.CreatorID != creatorID {
		return ds.Heating{}, fmt.Errorf("%w: вы не создатель этой заявки", ErrNotAllowed)
	}

	items, err := r.GetHeatingItems(int(heating.HeatingID))
	if err != nil {
		return ds.Heating{}, err
	}
	if len(items) == 0 {
		return ds.Heating{}, errors.New("нельзя сформировать пустую заявку")
	}

	for _, item := range items {
		var component ds.Component
		if err := r.db.First(&component, item.ComponentID).Error; err != nil {
			return ds.Heating{}, err
		}
		rt := CalculateHeat(
			heating.AmbientTemperature,
			component.ThermalResistance,
			item.PowerDissipation,
		)
		r.db.Model(&ds.HeatingComponent{}).
			Where("heating_id = ? AND component_id = ?", heating.HeatingID, item.ComponentID).
			Update("heat", rt)
	}

	formingDate := time.Now()
	err = r.db.Model(&heating).Updates(map[string]interface{}{
		"status":       "formed",
		"forming_date": formingDate,
	}).Error
	if err != nil {
		return ds.Heating{}, err
	}
	heating.Status = "formed"
	heating.FormingDate = &formingDate
	return heating, nil
}

func (r *Repository) FinishHeating(id int, status string, moderatorID uint) (ds.Heating, error) {
	if status != "completed" && status != "rejected" {
		return ds.Heating{}, errors.New("неверный статус: допустимы completed или rejected")
	}
	load, err := r.GetSingleHeating(id)
	if err != nil {
		return ds.Heating{}, err
	}
	if load.Status != "formed" {
		return ds.Heating{}, fmt.Errorf("%w: завершить/отклонить можно только сформированную заявку", ErrNotAllowed)
	}
	finishDate := time.Now()
	err = r.db.Model(&load).Updates(map[string]interface{}{
		"status":       status,
		"finish_date":  finishDate,
		"moderator_id": moderatorID,
	}).Error
	if err != nil {
		return ds.Heating{}, err
	}
	load.Status = status
	load.FinishDate = sql.NullTime{Time: finishDate, Valid: true}
	load.ModeratorID = &moderatorID
	return load, nil
}

func (r *Repository) DeleteHeating(heatingID int, creatorID uint) (ds.Heating, error) {
	heating, err := r.GetSingleHeating(heatingID)
	if err != nil {
		return ds.Heating{}, err
	}
	if heating.Status != "draft" {
		return ds.Heating{}, fmt.Errorf("%w: удалить можно только черновик", ErrNotAllowed)
	}
	if heating.CreatorID != creatorID {
		return ds.Heating{}, fmt.Errorf("%w: вы не создатель этой заявки", ErrNotAllowed)
	}
	formingDate := time.Now()
	err = r.db.Model(&heating).Updates(map[string]interface{}{
		"status":       "deleted",
		"forming_date": formingDate,
	}).Error
	if err != nil {
		return ds.Heating{}, err
	}
	heating.Status = "deleted"
	heating.FormingDate = &formingDate
	return heating, nil
}
