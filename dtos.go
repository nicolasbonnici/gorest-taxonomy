package taxonomy

import (
	"time"

	"github.com/google/uuid"
)

type CategoryCreateDTO struct {
	ParentID    *string `json:"parent_id,omitempty"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug,omitempty"`
	Description string  `json:"description,omitempty"`
}

type CategoryUpdateDTO struct {
	ParentID    *string `json:"parent_id,omitempty"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug,omitempty"`
	Description string  `json:"description,omitempty"`
}

type CategoryResponseDTO struct {
	ID          uuid.UUID  `json:"id"`
	Parent      *string    `json:"parent,omitempty"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description string     `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type CategoryTreeNode struct {
	ID          uuid.UUID           `json:"id"`
	ParentID    *uuid.UUID          `json:"parent_id,omitempty"`
	Name        string              `json:"name"`
	Slug        string              `json:"slug"`
	Description string              `json:"description,omitempty"`
	Children    []*CategoryTreeNode `json:"children,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   *time.Time          `json:"updated_at,omitempty"`
}

type TagCreateDTO struct {
	Name string `json:"name"`
	Slug string `json:"slug,omitempty"`
}

type TagUpdateDTO struct {
	Name string `json:"name"`
	Slug string `json:"slug,omitempty"`
}

type TagResponseDTO struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Slug      string     `json:"slug"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type ResourceAttachDTO struct {
	Resource   string `json:"resource"`
	ResourceID string `json:"resource_id"`
}
