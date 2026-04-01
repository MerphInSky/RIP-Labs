package serializer

import "web_backend/internal/app/ds"

type ComponentJSON struct {
	ComponentID       uint    `json:"component_id"`
	Title             string  `json:"title"`
	Description       string  `json:"description"`
	IsDeleted         bool    `json:"is_deleted"`
	PhotoURL          string  `json:"photo_url"`
	Video             string  `json:"video"`
	ThermalResistance float64 `json:"thermal_resistance"`
}

func ComponentToJSON(s ds.Component) ComponentJSON {
	return ComponentJSON{
		ComponentID:       s.ComponentID,
		Title:             s.Title,
		Description:       s.Description,
		IsDeleted:         s.IsDeleted,
		PhotoURL:          s.PhotoURL,
		Video:             s.Video,
		ThermalResistance: s.ThermalResistance,
	}
}

func ComponentFromJSON(j ComponentJSON) ds.Component {
	return ds.Component{
		Title:             j.Title,
		Description:       j.Description,
		ThermalResistance: j.ThermalResistance,
		Video:             j.Video,
	}
}
