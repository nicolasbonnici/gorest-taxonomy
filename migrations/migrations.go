package migrations

import (
	"context"

	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/migrations"
)

func GetMigrations() migrations.MigrationSource {
	builder := migrations.NewMigrationBuilder("gorest-taxonomy")
	builder.Add("20260516000001000", "create_categories_table", upCategories, downCategories)
	builder.Add("20260516000002000", "create_category_resources_table", upCategoryResources, downCategoryResources)
	builder.Add("20260516000003000", "create_tags_table", upTags, downTags)
	builder.Add("20260516000004000", "create_tag_resources_table", upTagResources, downTagResources)
	builder.Add("20260705000005000", "add_taxonomy_covering_indexes", upTaxonomyCoveringIndexes, downTaxonomyCoveringIndexes)
	return builder.Build()
}

func upCategories(ctx context.Context, db database.Database) error {
	if err := migrations.SQL(ctx, db, migrations.DialectSQL{
		Postgres: `CREATE TABLE IF NOT EXISTS categories (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(255) NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP(0) WITH TIME ZONE
		)`,
		MySQL: `CREATE TABLE IF NOT EXISTS categories (
			id CHAR(36) PRIMARY KEY,
			parent_id CHAR(36),
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(255) NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NULL,
			FOREIGN KEY (parent_id) REFERENCES categories(id) ON DELETE SET NULL,
			INDEX idx_categories_parent (parent_id),
			INDEX idx_categories_slug (slug)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		SQLite: `CREATE TABLE IF NOT EXISTS categories (
			id TEXT PRIMARY KEY,
			parent_id TEXT REFERENCES categories(id) ON DELETE SET NULL,
			name TEXT NOT NULL,
			slug TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at TEXT
		)`,
	}); err != nil {
		return err
	}
	if db.DriverName() != "mysql" {
		if err := migrations.CreateIndex(ctx, db, "idx_categories_parent", "categories", "parent_id"); err != nil {
			return err
		}
		return migrations.CreateIndex(ctx, db, "idx_categories_slug", "categories", "slug")
	}
	return nil
}

func downCategories(ctx context.Context, db database.Database) error {
	if db.DriverName() != "mysql" {
		_ = migrations.DropIndex(ctx, db, "idx_categories_parent", "categories")
		_ = migrations.DropIndex(ctx, db, "idx_categories_slug", "categories")
	}
	return migrations.DropTableIfExists(ctx, db, "categories")
}

func upCategoryResources(ctx context.Context, db database.Database) error {
	if err := migrations.SQL(ctx, db, migrations.DialectSQL{
		Postgres: `CREATE TABLE IF NOT EXISTS category_resources (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
			resource VARCHAR(255) NOT NULL,
			resource_id UUID NOT NULL,
			created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(category_id, resource, resource_id)
		)`,
		MySQL: `CREATE TABLE IF NOT EXISTS category_resources (
			id CHAR(36) PRIMARY KEY,
			category_id CHAR(36) NOT NULL,
			resource VARCHAR(255) NOT NULL,
			resource_id CHAR(36) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE,
			UNIQUE KEY uq_category_resource (category_id, resource, resource_id),
			INDEX idx_category_resources_lookup (resource, resource_id),
			INDEX idx_category_resources_category (category_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		SQLite: `CREATE TABLE IF NOT EXISTS category_resources (
			id TEXT PRIMARY KEY,
			category_id TEXT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
			resource TEXT NOT NULL,
			resource_id TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			UNIQUE(category_id, resource, resource_id)
		)`,
	}); err != nil {
		return err
	}
	if db.DriverName() != "mysql" {
		if err := migrations.SQL(ctx, db, migrations.DialectSQL{
			Postgres: `CREATE INDEX IF NOT EXISTS idx_category_resources_lookup ON category_resources(resource, resource_id)`,
			SQLite:   `CREATE INDEX IF NOT EXISTS idx_category_resources_lookup ON category_resources(resource, resource_id)`,
		}); err != nil {
			return err
		}
		return migrations.CreateIndex(ctx, db, "idx_category_resources_category", "category_resources", "category_id")
	}
	return nil
}

func downCategoryResources(ctx context.Context, db database.Database) error {
	if db.DriverName() != "mysql" {
		_ = migrations.DropIndex(ctx, db, "idx_category_resources_lookup", "category_resources")
		_ = migrations.DropIndex(ctx, db, "idx_category_resources_category", "category_resources")
	}
	return migrations.DropTableIfExists(ctx, db, "category_resources")
}

func upTags(ctx context.Context, db database.Database) error {
	if err := migrations.SQL(ctx, db, migrations.DialectSQL{
		Postgres: `CREATE TABLE IF NOT EXISTS tags (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(255) NOT NULL,
			created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP(0) WITH TIME ZONE
		)`,
		MySQL: `CREATE TABLE IF NOT EXISTS tags (
			id CHAR(36) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NULL,
			INDEX idx_tags_slug (slug)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		SQLite: `CREATE TABLE IF NOT EXISTS tags (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			slug TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at TEXT
		)`,
	}); err != nil {
		return err
	}
	if db.DriverName() != "mysql" {
		return migrations.CreateIndex(ctx, db, "idx_tags_slug", "tags", "slug")
	}
	return nil
}

func downTags(ctx context.Context, db database.Database) error {
	if db.DriverName() != "mysql" {
		_ = migrations.DropIndex(ctx, db, "idx_tags_slug", "tags")
	}
	return migrations.DropTableIfExists(ctx, db, "tags")
}

func upTagResources(ctx context.Context, db database.Database) error {
	if err := migrations.SQL(ctx, db, migrations.DialectSQL{
		Postgres: `CREATE TABLE IF NOT EXISTS tag_resources (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
			resource VARCHAR(255) NOT NULL,
			resource_id UUID NOT NULL,
			created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(tag_id, resource, resource_id)
		)`,
		MySQL: `CREATE TABLE IF NOT EXISTS tag_resources (
			id CHAR(36) PRIMARY KEY,
			tag_id CHAR(36) NOT NULL,
			resource VARCHAR(255) NOT NULL,
			resource_id CHAR(36) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE,
			UNIQUE KEY uq_tag_resource (tag_id, resource, resource_id),
			INDEX idx_tag_resources_lookup (resource, resource_id),
			INDEX idx_tag_resources_tag (tag_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		SQLite: `CREATE TABLE IF NOT EXISTS tag_resources (
			id TEXT PRIMARY KEY,
			tag_id TEXT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
			resource TEXT NOT NULL,
			resource_id TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			UNIQUE(tag_id, resource, resource_id)
		)`,
	}); err != nil {
		return err
	}
	if db.DriverName() != "mysql" {
		if err := migrations.SQL(ctx, db, migrations.DialectSQL{
			Postgres: `CREATE INDEX IF NOT EXISTS idx_tag_resources_lookup ON tag_resources(resource, resource_id)`,
			SQLite:   `CREATE INDEX IF NOT EXISTS idx_tag_resources_lookup ON tag_resources(resource, resource_id)`,
		}); err != nil {
			return err
		}
		return migrations.CreateIndex(ctx, db, "idx_tag_resources_tag", "tag_resources", "tag_id")
	}
	return nil
}

func downTagResources(ctx context.Context, db database.Database) error {
	if db.DriverName() != "mysql" {
		_ = migrations.DropIndex(ctx, db, "idx_tag_resources_lookup", "tag_resources")
		_ = migrations.DropIndex(ctx, db, "idx_tag_resources_tag", "tag_resources")
	}
	return migrations.DropTableIfExists(ctx, db, "tag_resources")
}
