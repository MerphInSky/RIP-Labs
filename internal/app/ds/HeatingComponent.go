package ds

type HeatingComponent struct {
	HeatingID uint `gorm:"primaryKey;column:heating_id"`
	ComponentID   uint `gorm:"primaryKey;column:component_id"`

	PowerDissipation   float64      `gorm:"type:numeric(5,2);not null;default:5.0"`
	Heat *float64 `gorm:"type:numeric(12,2)"`

	Component   Component `gorm:"foreignKey:ComponentID"`
	Heating Heating       `gorm:"foreignKey:HeatingID"`
}

func (HeatingComponent) TableName() string {
	return "heating_components"
}
