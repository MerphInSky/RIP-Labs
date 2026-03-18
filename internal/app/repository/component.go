package repository

import (
	"fmt"
	"web_backend/internal/app/ds"
)

func (r *Repository) GetComponents() ([]ds.Component, error) {
	var components []ds.Component
	err := r.db.Where("is_deleted = ?", false).Find(&components).Error
	if err != nil {
		return nil, err
	}
	if len(components) == 0 {
		return nil, fmt.Errorf("пусто")
	}
	return components, nil
}

func (r *Repository) GetComponent(id int) (ds.Component, error) {
	var component ds.Component
	err := r.db.Where("component_id = ? AND is_deleted = ?", id, false).First(&component).Error
	if err != nil {
		return ds.Component{}, err
	}
	return component, nil
}

func (r *Repository) GetComponentsByTitle(title string) ([]ds.Component, error) {
	var components []ds.Component
	err := r.db.Where("title ILIKE ? AND is_deleted = ?", "%"+title+"%", false).Find(&components).Error
	if err != nil {
		return nil, err
	}
	return components, nil
}
