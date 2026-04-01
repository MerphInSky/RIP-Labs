package ds

import (
	"database/sql"
	"time"
)

type Heating struct {
	HeatingID uint                `gorm:"primaryKey;column:heating_id"`
	Status       string           `gorm:"type:varchar(20);not null"`
	CreatedAt    time.Time        `gorm:"not null"`
	CreatorID    uint             `gorm:"not null"`
	FormingDate  *time.Time       `gorm:"column:forming_date"`
	FinishDate   sql.NullTime     `gorm:"column:finish_date"`
	ModeratorID  *uint            `gorm:"column:moderator_id"`
	Description  *string          `gorm:"type:varchar(2000)"`
	AmbientTemperature float64    `gorm:"type:numeric(5,2);not null;default:25.0"`
	
	Creator   Users  `gorm:"foreignKey:CreatorID"`
	Moderator *Users `gorm:"foreignKey:ModeratorID"`
}

func (Heating) TableName() string {
	return "heatings"
}
