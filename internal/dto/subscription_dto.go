package dto

import "github.com/google/uuid"

type CreateSubscriptionDTO struct {
	ServiceName string    `json:"service_name" example:"Yandex Plus"`
	Price       int       `json:"price" example:"400"`
	UserID      uuid.UUID `json:"user_id" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	StartDate   string    `json:"start_date" example:"07-2025"`
	EndDate     string    `json:"end_date,omitempty" example:"12-2025"` // optional
}

type UpdateSubscriptionDTO struct {
	ServiceName *string    `json:"service_name,omitempty" example:"Netflix"`
	Price       *int       `json:"price,omitempty" example:"500"`
	StartDate   *string    `json:"start_date,omitempty" example:"08-2025"`
	EndDate     *string    `json:"end_date,omitempty" example:"12-2025"`
	UserID      *uuid.UUID `json:"user_id,omitempty" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
}

// Ответ при создании
type CreateSubscriptionResponse struct {
	ID int `json:"id" example:"1"`
}
