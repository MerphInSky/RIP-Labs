package repository

import (
	"errors"
	"fmt"
	"time"
	"web_backend/internal/app/ds"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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

func (r *Repository) GetHeating(id int, creatorID uint) ([]ds.HeatingComponent, *ds.Heating, error) {
	var load ds.Heating
	err := r.db.Where("heating_id = ? AND creator_id = ? AND status != ?",
		id, creatorID, "deleted").First(&load).Error
	if err != nil {
		return nil, nil, err
	}

	var items []ds.HeatingComponent
	err = r.db.Where("heating_id = ?", id).
		Preload("Component").
		Find(&items).Error
	if err != nil {
		return nil, nil, err
	}
	return items, &load, nil
}

func (r *Repository) AddComponent(componentID uint, creatorID uint) error {
	var load ds.Heating

	err := r.db.Where("creator_id = ? AND status = ?", creatorID, "draft").
		First(&load).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		load = ds.Heating{
			Status:    "draft",
			CreatedAt: time.Now(),
			CreatorID: creatorID,
			AmbientTemperature: 25,
		}
		if err := r.db.Create(&load).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	var count int64
	r.db.Model(&ds.HeatingComponent{}).
		Where("heating_id = ? AND component_id = ?", load.HeatingID, componentID).
		Count(&count)

	if count == 0 {
		var component ds.Component
		if err := r.db.First(&component, componentID).Error; err != nil {
			return err
		}

		heat := CalculateHeat(
			25,component.ThermalResistance, 5,
		)

		item := ds.HeatingComponent{
			HeatingID: load.HeatingID,
			ComponentID:   componentID,
			PowerDissipation:   1000,
			Heat: &heat,
		}
		if err := r.db.Create(&item).Error; err != nil {
			return err
		}
	}

	return nil
}

// без ORM
func (r *Repository) DeleteHeating(loadID uint) error {
	query := `
		UPDATE heatings
		SET status = 'deleted'
		WHERE heating_id = $1;
	`
	result := r.db.Exec(query, loadID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("heating with id %d not found", loadID)
	}
	return nil
}

func (r *Repository) IsDraftHeating(loadID int, creatorID uint) (bool, error) {
	var load ds.Heating
	err := r.db.Select("status").Where("heating_id = ? AND creator_id = ?",
		loadID, creatorID).First(&load).Error
	if err != nil {
		return false, err
	}
	return load.Status == "draft", nil
}
