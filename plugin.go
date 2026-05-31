package taxonomy

import (
	"github.com/gofiber/fiber/v3"
	"github.com/nicolasbonnici/gorest-taxonomy/migrations"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/logger"
	"github.com/nicolasbonnici/gorest/plugin"
)

type TaxonomyPlugin struct {
	config Config
	db     database.Database
}

func NewPlugin() plugin.Plugin {
	return &TaxonomyPlugin{}
}

func (p *TaxonomyPlugin) Name() string {
	return "taxonomy"
}

func (p *TaxonomyPlugin) Initialize(config map[string]interface{}) error {
	p.config = DefaultConfig()

	if db, ok := config["database"].(database.Database); ok {
		p.db = db
		p.config.Database = db
	}

	if allowedTypes, ok := config["allowed_types"].([]interface{}); ok {
		types := make([]string, 0, len(allowedTypes))
		for _, t := range allowedTypes {
			if str, ok := t.(string); ok {
				types = append(types, str)
			}
		}
		if len(types) > 0 {
			p.config.AllowedTypes = types
		}
	}

	if maxDepth, ok := config["max_depth"].(int); ok {
		p.config.MaxDepth = maxDepth
	}

	if paginationLimit, ok := config["pagination_limit"].(int); ok {
		p.config.PaginationLimit = paginationLimit
	}

	if maxPaginationLimit, ok := config["max_pagination_limit"].(int); ok {
		p.config.MaxPaginationLimit = maxPaginationLimit
	}

	return p.config.Validate()
}

func (p *TaxonomyPlugin) Handler() fiber.Handler {
	return func(c fiber.Ctx) error {
		return c.Next()
	}
}

func (p *TaxonomyPlugin) SetupEndpoints(router fiber.Router) error {
	if p.db == nil {
		logger.Log.Warn("Taxonomy plugin database not initialized, skipping endpoint registration")
		return nil
	}

	RegisterRoutes(router, p.db, &p.config)

	logger.Log.Info("Taxonomy plugin endpoints registered")
	return nil
}

func (p *TaxonomyPlugin) MigrationSource() interface{} {
	return migrations.GetMigrations()
}

func (p *TaxonomyPlugin) Dependencies() []string {
	return []string{}
}

func (p *TaxonomyPlugin) MigrationDependencies() []string {
	return []string{}
}

func (p *TaxonomyPlugin) GetOpenAPIResources() []plugin.OpenAPIResource {
	return []plugin.OpenAPIResource{
		{
			Name:          "category",
			PluralName:    "categories",
			BasePath:      "/categories",
			Tags:          []string{"Taxonomy"},
			ResponseModel: CategoryResponseDTO{},
			CreateModel:   CategoryCreateDTO{},
			UpdateModel:   CategoryUpdateDTO{},
			Description:   "Hierarchical categories for any resource type",
		},
		{
			Name:          "tag",
			PluralName:    "tags",
			BasePath:      "/tags",
			Tags:          []string{"Taxonomy"},
			ResponseModel: TagResponseDTO{},
			CreateModel:   TagCreateDTO{},
			UpdateModel:   TagUpdateDTO{},
			Description:   "Flat tags for any resource type",
		},
	}
}
