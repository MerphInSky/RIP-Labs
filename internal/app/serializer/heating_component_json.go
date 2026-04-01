package serializer

import "web_backend/internal/app/ds"

type HeatingComponentJSON struct {
	HeatingID uint     `json:"heating_id"`
	ComponentID   uint     `json:"component_id"`
	PowerDissipation   float64      `json:"power_dissipation"`
	Heat *float64	`json:"heat"`
}


type HeatingComponentDetailJSON struct {
	HeatingID uint                 `json:"heating_id"`
	ComponentID   uint                 `json:"component_id"`
	PowerDissipation   float64                  `json:"power_dissipation"`
	Component     ComponentJSON `json:"component"`
	Heat *float64	`json:"heat"`
}

func HeatingComponentToJSON(item ds.HeatingComponent) HeatingComponentJSON {
	return HeatingComponentJSON{
		HeatingID: item.HeatingID,
		ComponentID:   item.ComponentID,
		PowerDissipation:   item.PowerDissipation,
		Heat: item.Heat,
	}
}

func HeatingComponentDetailToJSON(item ds.HeatingComponent) HeatingComponentDetailJSON {
	return HeatingComponentDetailJSON{
		HeatingID: item.HeatingID,
		ComponentID:   item.ComponentID,
		PowerDissipation:   item.PowerDissipation,
		Heat: item.Heat,
		Component:     ComponentToJSON(item.Component),
	}
}

func HeatingComponentFromJSON(j HeatingComponentJSON) ds.HeatingComponent {
	return ds.HeatingComponent{
		PowerDissipation: j.PowerDissipation,
	}
}
