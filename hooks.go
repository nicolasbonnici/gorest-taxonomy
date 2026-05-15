package taxonomy

import (
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	auth "github.com/nicolasbonnici/gorest/auth"
	"github.com/nicolasbonnici/gorest/crud"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/query"
)

var nonAlphanumRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	return strings.Trim(nonAlphanumRe.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

type TaxonomyHooks struct {
	db      database.Database
	config  *Config
	service *TaxonomyService
}

func NewTaxonomyHooks(db database.Database, config *Config, service *TaxonomyService) *TaxonomyHooks {
	return &TaxonomyHooks{db: db, config: config, service: service}
}

func (h *TaxonomyHooks) CategoryCreateHook(c *fiber.Ctx, dto CategoryCreateDTO, model *Category) error {
	name := strings.TrimSpace(dto.Name)
	if name == "" {
		return fiber.NewError(400, "name is required")
	}
	model.Name = name

	if dto.Slug != "" {
		model.Slug = dto.Slug
	} else {
		model.Slug = slugify(name)
	}

	if dto.ParentID != nil {
		parentUUID, err := uuid.Parse(*dto.ParentID)
		if err != nil {
			return fiber.NewError(400, "invalid parent_id")
		}
		depth, err := h.service.GetCategoryDepth(auth.Context(c), parentUUID)
		if err != nil {
			return fiber.NewError(400, "invalid parent_id")
		}
		if depth >= h.config.MaxDepth-1 {
			return fiber.NewError(400, "maximum category depth exceeded")
		}
		model.ParentID = &parentUUID
	}

	model.CreatedAt = time.Now()
	return nil
}

func (h *TaxonomyHooks) CategoryUpdateHook(c *fiber.Ctx, dto CategoryUpdateDTO, model *Category) error {
	name := strings.TrimSpace(dto.Name)
	if name == "" {
		return fiber.NewError(400, "name is required")
	}
	model.Name = name

	if dto.Slug != "" {
		model.Slug = dto.Slug
	} else {
		model.Slug = slugify(name)
	}

	model.Description = dto.Description

	if dto.ParentID != nil {
		parentUUID, err := uuid.Parse(*dto.ParentID)
		if err != nil {
			return fiber.NewError(400, "invalid parent_id")
		}
		depth, err := h.service.GetCategoryDepth(auth.Context(c), parentUUID)
		if err != nil {
			return fiber.NewError(400, "invalid parent_id")
		}
		if depth >= h.config.MaxDepth-1 {
			return fiber.NewError(400, "maximum category depth exceeded")
		}
		model.ParentID = &parentUUID
	}

	now := time.Now()
	model.UpdatedAt = &now
	return nil
}

func (h *TaxonomyHooks) CategoryDeleteHook(_ *fiber.Ctx, _ any) error {
	return nil
}

func (h *TaxonomyHooks) CategoryGetByIDHook(_ *fiber.Ctx, _ any) error {
	return nil
}

func (h *TaxonomyHooks) CategoryGetAllHook(_ *fiber.Ctx, _ *[]query.Condition, _ *[]crud.OrderByClause) error {
	return nil
}

func (h *TaxonomyHooks) TagCreateHook(c *fiber.Ctx, dto TagCreateDTO, model *Tag) error {
	name := strings.TrimSpace(dto.Name)
	if name == "" {
		return fiber.NewError(400, "name is required")
	}
	model.Name = name

	if dto.Slug != "" {
		model.Slug = dto.Slug
	} else {
		model.Slug = slugify(name)
	}

	model.CreatedAt = time.Now()
	return nil
}

func (h *TaxonomyHooks) TagUpdateHook(_ *fiber.Ctx, dto TagUpdateDTO, model *Tag) error {
	name := strings.TrimSpace(dto.Name)
	if name == "" {
		return fiber.NewError(400, "name is required")
	}
	model.Name = name

	if dto.Slug != "" {
		model.Slug = dto.Slug
	} else {
		model.Slug = slugify(name)
	}

	now := time.Now()
	model.UpdatedAt = &now
	return nil
}

func (h *TaxonomyHooks) TagDeleteHook(_ *fiber.Ctx, _ any) error {
	return nil
}

func (h *TaxonomyHooks) TagGetByIDHook(_ *fiber.Ctx, _ any) error {
	return nil
}

func (h *TaxonomyHooks) TagGetAllHook(_ *fiber.Ctx, _ *[]query.Condition, _ *[]crud.OrderByClause) error {
	return nil
}
