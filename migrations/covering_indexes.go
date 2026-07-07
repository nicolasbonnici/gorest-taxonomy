package migrations

import (
	"context"

	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/migrations"
)

// The base (resource, resource_id) composite index already serves the
// polymorphic lookup filter. Extending it with the join key (category_id /
// tag_id) lets Postgres satisfy the batch category/tag load with an index-only
// scan, avoiding a heap fetch per matched pivot row. Plain multi-column indexes
// are used (not Postgres INCLUDE) so the same DDL runs on SQLite in tests.
func upTaxonomyCoveringIndexes(ctx context.Context, db database.Database) error {
	if err := migrations.CreateIndex(ctx, db, "idx_category_resources_covering", "category_resources", "resource, resource_id, category_id"); err != nil {
		return err
	}
	return migrations.CreateIndex(ctx, db, "idx_tag_resources_covering", "tag_resources", "resource, resource_id, tag_id")
}

func downTaxonomyCoveringIndexes(ctx context.Context, db database.Database) error {
	_ = migrations.DropIndex(ctx, db, "idx_category_resources_covering", "category_resources")
	return migrations.DropIndex(ctx, db, "idx_tag_resources_covering", "tag_resources")
}
