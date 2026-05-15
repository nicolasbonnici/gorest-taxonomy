package taxonomy

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`
	Name        string     `json:"name" db:"name"`
	Slug        string     `json:"slug" db:"slug"`
	Description string     `json:"description" db:"description"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

func (Category) TableName() string { return "categories" }

type CategoryResource struct {
	ID         uuid.UUID `json:"id" db:"id"`
	CategoryID uuid.UUID `json:"category_id" db:"category_id"`
	Resource   string    `json:"resource" db:"resource"`
	ResourceID uuid.UUID `json:"resource_id" db:"resource_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

func (CategoryResource) TableName() string { return "category_resources" }

type Tag struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	Slug      string     `json:"slug" db:"slug"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

func (Tag) TableName() string { return "tags" }

type TagResource struct {
	ID         uuid.UUID `json:"id" db:"id"`
	TagID      uuid.UUID `json:"tag_id" db:"tag_id"`
	Resource   string    `json:"resource" db:"resource"`
	ResourceID uuid.UUID `json:"resource_id" db:"resource_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

func (TagResource) TableName() string { return "tag_resources" }
