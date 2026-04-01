package serializer

import (
	"time"

	"web_backend/internal/app/ds"
)

type HeatingJSON struct {
	HeatingID       uint       `json:"heating_id"`
	Status             string     `json:"status"`
	CreatedAt          time.Time  `json:"created_at"`
	CreatorLogin       string     `json:"creator_login"`
	ModeratorLogin     *string    `json:"moderator_login"`
	FormingDate        *time.Time `json:"forming_date"`
	FinishDate         *time.Time `json:"finish_date"`
	Description        *string    `json:"description"`
	CompletedItemCount int      `json:"completed_item_count"`
	AmbientTemperature float64    `json:"ambient_temperature"`
}

func HeatingToJSON(app ds.Heating, creatorLogin, moderatorLogin string, completedItemCount int) HeatingJSON {
	var mLogin *string
	if moderatorLogin != "" {
		mLogin = &moderatorLogin
	}
	var formingDate, finishDate *time.Time
	if app.FormingDate != nil {
		formingDate = app.FormingDate
	}
	if app.FinishDate.Valid {
		finishDate = &app.FinishDate.Time
	}
	return HeatingJSON{
		HeatingID:            app.HeatingID,
		Status:               app.Status,
		CreatedAt:            app.CreatedAt,
		CreatorLogin:         creatorLogin,
		ModeratorLogin:       mLogin,
		FormingDate:          formingDate,
		FinishDate:           finishDate,
		CompletedItemCount:   completedItemCount,
	}
}

func HeatingFromJSON(j HeatingJSON) ds.Heating {
	return ds.Heating{
		AmbientTemperature: j.AmbientTemperature,
	}
}


type StatusJSON struct {
	Status string `json:"status"`
}
