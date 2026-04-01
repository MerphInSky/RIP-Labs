package repository

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"web_backend/internal/app/ds"
	minioClient "web_backend/internal/app/minioClient"
	"web_backend/internal/app/serializer"
)

func (r *Repository) GetComponents() ([]ds.Component, error) {
	var components []ds.Component
	err := r.db.Where("is_deleted = ?", false).Find(&components).Error
	if err != nil {
		return nil, err
	}
	return components, nil
}

func (r *Repository) GetComponent(id int) (*ds.Component, error) {
	var component ds.Component
	err := r.db.Where("component_id = ? AND is_deleted = ?", id, false).First(&component).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: стратегия с id %d", ErrNotFound, id)
		}
		return nil, err
	}
	return &component, nil
}

func (r *Repository) GetComponentsByTitle(title string) ([]ds.Component, error) {
	var components []ds.Component
	err := r.db.Where("title ILIKE ? AND is_deleted = ?", "%"+title+"%", false).Find(&components).Error
	if err != nil {
		return nil, err
	}
	return components, nil
}

func (r *Repository) CreateComponent(j serializer.ComponentJSON) (ds.Component, error) {
	component := serializer.ComponentFromJSON(j)
	err := r.db.Create(&component).Scan(&component).Error
	if err != nil {
		return ds.Component{}, err
	}
	return component, nil
}

func (r *Repository) AddPhoto(ctx *gin.Context, componentID int, file *multipart.FileHeader) (*ds.Component, error) {
	component, err := r.GetComponent(componentID)
	if err != nil {
		return nil, err
	}
	if component.PhotoURL != "" {
		_ = minioClient.DeleteObject(ctx, r.mc, minioClient.GetImgBucket(), component.PhotoURL)
	}
	fileName, err := minioClient.UploadImage(ctx, r.mc, minioClient.GetImgBucket(), file, component.ComponentID)
	if err != nil {
		return nil, err
	}
	if err := r.db.Model(&ds.Component{}).Where("component_id = ?", componentID).Update("photo_url", fileName).Error; err != nil {
		return nil, err
	}
	component.PhotoURL = fileName
	return component, nil
}

func (r *Repository) AddVideo(ctx *gin.Context, componentID int, file *multipart.FileHeader) (*ds.Component, error) {
	component, err := r.GetComponent(componentID)
	if err != nil {
		return nil, err
	}
	if component.Video != "" {
		_ = minioClient.DeleteObject(ctx, r.mc, minioClient.GetImgBucket(), component.Video)
	}
	fileName, err := minioClient.UploadVideo(ctx, r.mc, minioClient.GetImgBucket(), file, component.ComponentID)
	if err != nil {
		return nil, err
	}
	if err := r.db.Model(&ds.Component{}).Where("component_id = ?", componentID).Update("video", fileName).Error; err != nil {
		return nil, err
	}
	component.Video = fileName
	return component, nil
}

func (r *Repository) DeleteComponent(id int) error {
	component, err := r.GetComponent(id)
	if err != nil {
		return err
	}
	if component.PhotoURL != "" {
		_ = minioClient.DeleteObject(context.Background(), r.mc, minioClient.GetImgBucket(), component.PhotoURL)
	}
	if component.Video != "" {
		_ = minioClient.DeleteObject(context.Background(), r.mc, minioClient.GetImgBucket(), component.Video)
	}
	return r.db.Model(&ds.Component{}).Where("component_id = ?", id).Update("is_deleted", true).Error
}
