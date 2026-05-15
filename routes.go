package taxonomy

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nicolasbonnici/gorest/database"
)

func RegisterRoutes(router fiber.Router, db database.Database, config *Config) {
	service := NewTaxonomyService(db, config)
	RegisterCategoryRoutes(router, db, config, service)
	RegisterTagRoutes(router, db, config, service)
	registerResourceLookupRoutes(router, config, service)
}

func registerResourceLookupRoutes(router fiber.Router, config *Config, service *TaxonomyService) {
	catConverter := &CategoryConverter{}
	tagConverter := &TagConverter{}

	router.Get("/:resource/:resource_id/categories", func(c *fiber.Ctx) error {
		resource := c.Params("resource")
		if !config.IsAllowedType(resource) {
			return fiber.NewError(400, "resource type is not allowed")
		}
		resourceID, err := uuid.Parse(c.Params("resource_id"))
		if err != nil {
			return fiber.NewError(400, "invalid resource_id")
		}
		categories, err := service.GetCategoriesForResource(c.Context(), resource, resourceID)
		if err != nil {
			return fiber.NewError(500, "failed to fetch categories")
		}
		return c.JSON(catConverter.ModelsToResponseDTOs(categories))
	})

	router.Get("/:resource/:resource_id/tags", func(c *fiber.Ctx) error {
		resource := c.Params("resource")
		if !config.IsAllowedType(resource) {
			return fiber.NewError(400, "resource type is not allowed")
		}
		resourceID, err := uuid.Parse(c.Params("resource_id"))
		if err != nil {
			return fiber.NewError(400, "invalid resource_id")
		}
		tags, err := service.GetTagsForResource(c.Context(), resource, resourceID)
		if err != nil {
			return fiber.NewError(500, "failed to fetch tags")
		}
		return c.JSON(tagConverter.ModelsToResponseDTOs(tags))
	})
}
