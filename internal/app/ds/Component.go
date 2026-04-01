package ds

type Component struct {
	ComponentID            uint    `gorm:"primaryKey;column:component_id"`
	Title                  string  `gorm:"type:varchar(255);not null"`
	Description            string  `gorm:"type:varchar(1000);not null"`
	IsDeleted              bool    `gorm:"type:boolean;not null;default:false"`
	PhotoURL               string  `gorm:"column:photo_url;type:varchar(255)"`
	Video                  string  `gorm:"type:varchar(255)"`
	ThermalResistance      float64 `gorm:"type:numeric(5,2);not null;default:1.0"`
}

func (Component) TableName() string {
	return "components"
}
