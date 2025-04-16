package model

import (
	"cakestore/internal/entity"
	"database/sql"
	"time"
)

type CakeModel struct {
	ID          int          `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Rating      int          `json:"rating"`
	ImageURL    string       `json:"image_url"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	DeletedAt   sql.NullTime `json:"deleted_at"`
}

type CreateUpdateCakeRequest struct {
	Title       string  `json:"title" validate:"required,min=3,max=100"`
	Description string  `json:"description" validate:"required"`
	Rating      float64 `json:"rating" validate:"required,gte=0,lte=10"`
	ImageURL    string  `json:"image" validate:"required,url"`
}

func CakeToResponse(cake *entity.Cake) *CakeModel {
	return &CakeModel{
		ID:          cake.ID,
		Title:       cake.Title,
		Description: cake.Description,
		Rating:      int(cake.Rating),
		ImageURL:    cake.Image,
		CreatedAt:   cake.CreatedAt,
		UpdatedAt:   cake.UpdatedAt,
		DeletedAt:   cake.DeletedAt,
	}
}
