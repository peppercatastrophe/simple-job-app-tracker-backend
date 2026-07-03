package models

import (
	"time"
)

type JobApplication struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	Title         string     `json:"title"`
	Company       string     `json:"company"`
	Category      *string    `json:"category,omitempty"`
	Description   *string    `json:"description,omitempty"`
	TechStack     []string   `json:"tech_stack,omitempty"`
	Status        string     `json:"status"`
	ApplyDateTime *time.Time `json:"apply_date_time,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type CreateJobApplicationInput struct {
	Title         string    `json:"title" validate:"required"`
	Company       string    `json:"company" validate:"required"`
	Category      *string   `json:"category,omitempty"`
	Description   *string   `json:"description,omitempty"`
	TechStack     []string  `json:"tech_stack,omitempty"`
	Status        string    `json:"status,omitempty"`
	ApplyDateTime *string   `json:"apply_date_time,omitempty"`
}

type UpdateJobApplicationInput struct {
	Title         *string   `json:"title,omitempty"`
	Company       *string   `json:"company,omitempty"`
	Category      *string   `json:"category,omitempty"`
	Description   *string   `json:"description,omitempty"`
	TechStack     []string  `json:"tech_stack,omitempty"`
	Status        *string   `json:"status,omitempty"`
	ApplyDateTime *string   `json:"apply_date_time,omitempty"`
}
