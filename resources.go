package taxonomy

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/nicolasbonnici/gorest/crud"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/processor"
	"github.com/nicolasbonnici/gorest/rbac"
)

type categoryHandler struct {
	processor processor.Processor[Category, CategoryCreateDTO, CategoryUpdateDTO, CategoryResponseDTO]
	service   *TaxonomyService
	config    *Config
}

type tagHandler struct {
	processor processor.Processor[Tag, TagCreateDTO, TagUpdateDTO, TagResponseDTO]
	service   *TaxonomyService
	config    *Config
}

// guards returns the middleware chain every mutating taxonomy route must carry:
// resolve the identity, load its roles, demand one, then demand a write role.
func guards(db database.Database, config *Config) []fiber.Handler {
	chain := []fiber.Handler{}
	if config.AuthMiddleware != nil {
		chain = append(chain, config.AuthMiddleware)
	}
	return append(chain,
		rbac.RoleLoader(db, config.RoleHierarchy),
		rbac.RequireAuthenticated(),
		rbac.RequireAnyRole(config.RoleHierarchy, config.SuperuserRole, config.WriteRoles...),
	)
}

// readChain mounts only the identity resolver, so a public read still knows who
// is asking when a token happens to be present.
func readChain(config *Config) []fiber.Handler {
	if config.AuthMiddleware != nil {
		return []fiber.Handler{config.AuthMiddleware}
	}
	return nil
}

// mount registers handler at the end of chain. Fiber v3 takes the first element
// positionally and the rest variadically, and runs them in the order given, so
// the guards have to be spliced in ahead of the handler here rather than by the
// caller.
func mount(register func(string, any, ...any) fiber.Router, path string, chain []fiber.Handler, h fiber.Handler) {
	all := make([]any, 0, len(chain)+1)
	for _, m := range chain {
		all = append(all, m)
	}
	all = append(all, h)
	register(path, all[0], all[1:]...)
}

func RegisterCategoryRoutes(router fiber.Router, db database.Database, config *Config, service *TaxonomyService) {
	hooks := NewTaxonomyHooks(db, config, service)

	proc := processor.New(processor.ProcessorConfig[Category, CategoryCreateDTO, CategoryUpdateDTO, CategoryResponseDTO]{
		DB:                 db,
		CRUD:               crud.New[Category](db),
		Converter:          &CategoryConverter{},
		PaginationLimit:    config.PaginationLimit,
		PaginationMaxLimit: config.MaxPaginationLimit,
		FieldMap: map[string]string{
			"id":          "id",
			"parent_id":   "parent_id",
			"name":        "name",
			"slug":        "slug",
			"description": "description",
			"created_at":  "created_at",
			"updated_at":  "updated_at",
		},
		AllowedFields: []string{"id", "parent_id", "name", "slug", "description", "created_at", "updated_at"},
	}).
		WithCreateHook(hooks.CategoryCreateHook).
		WithUpdateHook(hooks.CategoryUpdateHook).
		WithDeleteHook(hooks.CategoryDeleteHook).
		WithGetByIDHook(hooks.CategoryGetByIDHook).
		WithGetAllHook(hooks.CategoryGetAllHook)

	h := &categoryHandler{processor: proc, service: service, config: config}

	write := guards(db, config)
	read := readChain(config)

	mount(router.Get, "/categories/tree", read, h.GetTree)
	mount(router.Get, "/categories/:id", read, h.GetByID)
	mount(router.Get, "/categories", read, h.GetAll)
	mount(router.Post, "/categories", write, h.Create)
	mount(router.Put, "/categories/:id", write, h.Update)
	mount(router.Delete, "/categories/:id", write, h.Delete)
	mount(router.Post, "/categories/:id/resources", write, h.AttachResource)
	mount(router.Delete, "/categories/:id/resources/:resource/:resource_id", write, h.DetachResource)
}

func RegisterTagRoutes(router fiber.Router, db database.Database, config *Config, service *TaxonomyService) {
	hooks := NewTaxonomyHooks(db, config, service)

	proc := processor.New(processor.ProcessorConfig[Tag, TagCreateDTO, TagUpdateDTO, TagResponseDTO]{
		DB:                 db,
		CRUD:               crud.New[Tag](db),
		Converter:          &TagConverter{},
		PaginationLimit:    config.PaginationLimit,
		PaginationMaxLimit: config.MaxPaginationLimit,
		FieldMap: map[string]string{
			"id":         "id",
			"name":       "name",
			"slug":       "slug",
			"created_at": "created_at",
			"updated_at": "updated_at",
		},
		AllowedFields: []string{"id", "name", "slug", "created_at", "updated_at"},
	}).
		WithCreateHook(hooks.TagCreateHook).
		WithUpdateHook(hooks.TagUpdateHook).
		WithDeleteHook(hooks.TagDeleteHook).
		WithGetByIDHook(hooks.TagGetByIDHook).
		WithGetAllHook(hooks.TagGetAllHook)

	h := &tagHandler{processor: proc, service: service, config: config}

	write := guards(db, config)
	read := readChain(config)

	mount(router.Get, "/tags/:id", read, h.GetByID)
	mount(router.Get, "/tags", read, h.GetAll)
	mount(router.Post, "/tags", write, h.Create)
	mount(router.Put, "/tags/:id", write, h.Update)
	mount(router.Delete, "/tags/:id", write, h.Delete)
	mount(router.Post, "/tags/:id/resources", write, h.AttachResource)
	mount(router.Delete, "/tags/:id/resources/:resource/:resource_id", write, h.DetachResource)
}

func (h *categoryHandler) Create(c fiber.Ctx) error {
	err := h.processor.Create(c)
	if err == nil {
		h.service.InvalidateCategoryCache()
	}
	return err
}

func (h *categoryHandler) GetByID(c fiber.Ctx) error {
	return h.processor.GetByID(c)
}

func (h *categoryHandler) GetAll(c fiber.Ctx) error {
	return h.processor.GetAll(c)
}

func (h *categoryHandler) Update(c fiber.Ctx) error {
	err := h.processor.Update(c)
	if err == nil {
		h.service.InvalidateCategoryCache()
	}
	return err
}

func (h *categoryHandler) Delete(c fiber.Ctx) error {
	err := h.processor.Delete(c)
	if err == nil {
		h.service.InvalidateCategoryCache()
	}
	return err
}

func (h *categoryHandler) GetTree(c fiber.Ctx) error {
	categories, err := h.service.GetAllCategories(c.Context())
	if err != nil {
		return fiber.NewError(500, "failed to fetch categories")
	}
	return c.JSON(h.service.BuildCategoryTree(categories))
}

func (h *categoryHandler) AttachResource(c fiber.Ctx) error {
	categoryID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(400, "invalid category id")
	}

	var dto ResourceAttachDTO
	if err := c.Bind().Body(&dto); err != nil {
		return fiber.NewError(400, "invalid request body")
	}

	if !h.config.IsAllowedType(dto.Resource) {
		return fiber.NewError(400, "resource type is not allowed")
	}

	resourceID, err := uuid.Parse(dto.ResourceID)
	if err != nil {
		return fiber.NewError(400, "invalid resource_id")
	}

	if err := h.service.AttachCategory(c.Context(), categoryID, dto.Resource, resourceID); err != nil {
		return fiber.NewError(500, "failed to attach resource")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "resource attached"})
}

func (h *categoryHandler) DetachResource(c fiber.Ctx) error {
	categoryID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(400, "invalid category id")
	}

	resource := c.Params("resource")
	if !h.config.IsAllowedType(resource) {
		return fiber.NewError(400, "resource type is not allowed")
	}

	resourceID, err := uuid.Parse(c.Params("resource_id"))
	if err != nil {
		return fiber.NewError(400, "invalid resource_id")
	}

	if err := h.service.DetachCategory(c.Context(), categoryID, resource, resourceID); err != nil {
		return fiber.NewError(500, "failed to detach resource")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *tagHandler) Create(c fiber.Ctx) error {
	return h.processor.Create(c)
}

func (h *tagHandler) GetByID(c fiber.Ctx) error {
	return h.processor.GetByID(c)
}

func (h *tagHandler) GetAll(c fiber.Ctx) error {
	return h.processor.GetAll(c)
}

func (h *tagHandler) Update(c fiber.Ctx) error {
	return h.processor.Update(c)
}

func (h *tagHandler) Delete(c fiber.Ctx) error {
	return h.processor.Delete(c)
}

func (h *tagHandler) AttachResource(c fiber.Ctx) error {
	tagID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(400, "invalid tag id")
	}

	var dto ResourceAttachDTO
	if err := c.Bind().Body(&dto); err != nil {
		return fiber.NewError(400, "invalid request body")
	}

	if !h.config.IsAllowedType(dto.Resource) {
		return fiber.NewError(400, "resource type is not allowed")
	}

	resourceID, err := uuid.Parse(dto.ResourceID)
	if err != nil {
		return fiber.NewError(400, "invalid resource_id")
	}

	if err := h.service.AttachTag(c.Context(), tagID, dto.Resource, resourceID); err != nil {
		return fiber.NewError(500, "failed to attach resource")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "resource attached"})
}

func (h *tagHandler) DetachResource(c fiber.Ctx) error {
	tagID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(400, "invalid tag id")
	}

	resource := c.Params("resource")
	if !h.config.IsAllowedType(resource) {
		return fiber.NewError(400, "resource type is not allowed")
	}

	resourceID, err := uuid.Parse(c.Params("resource_id"))
	if err != nil {
		return fiber.NewError(400, "invalid resource_id")
	}

	if err := h.service.DetachTag(c.Context(), tagID, resource, resourceID); err != nil {
		return fiber.NewError(500, "failed to detach resource")
	}

	return c.SendStatus(fiber.StatusNoContent)
}
