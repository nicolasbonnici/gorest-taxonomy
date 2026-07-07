package taxonomy

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	pluginmigrations "github.com/nicolasbonnici/gorest-taxonomy/migrations"
	"github.com/nicolasbonnici/gorest/database"
	_ "github.com/nicolasbonnici/gorest/database/sqlite"
)

func setupServiceTestDB(t *testing.T) (database.Database, *TaxonomyService) {
	t.Helper()

	db, err := database.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory sqlite: %v", err)
	}

	ctx := context.Background()
	source := pluginmigrations.GetMigrations()
	migs, err := source.Migrations()
	if err != nil {
		t.Fatalf("load migrations: %v", err)
	}
	for _, m := range migs {
		if err := m.ExecuteUp(ctx, db); err != nil {
			t.Fatalf("apply migration %s: %v", m.FullName(), err)
		}
	}

	cfg := DefaultConfig()
	return db, NewTaxonomyService(db, &cfg)
}

func nowStamp() string { return time.Now().UTC().Format("2006-01-02 15:04:05") }

func insertCategory(t *testing.T, db database.Database, id uuid.UUID, parent *uuid.UUID, name string) {
	t.Helper()
	_, err := db.Exec(context.Background(),
		"INSERT INTO categories (id, parent_id, name, slug, description, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		id, parent, name, name, "", nowStamp())
	if err != nil {
		t.Fatalf("insert category: %v", err)
	}
}

func TestService_BatchLoadEliminatesN1(t *testing.T) {
	db, svc := setupServiceTestDB(t)
	ctx := context.Background()

	cat := uuid.New()
	insertCategory(t, db, cat, nil, "news")

	postA, postB, postC := uuid.New(), uuid.New(), uuid.New()
	if err := svc.AttachCategory(ctx, cat, "post", postA); err != nil {
		t.Fatalf("attach A: %v", err)
	}
	if err := svc.AttachCategory(ctx, cat, "post", postB); err != nil {
		t.Fatalf("attach B: %v", err)
	}

	byResource, err := svc.GetCategoriesForResources(ctx, "post", []uuid.UUID{postA, postB, postC})
	if err != nil {
		t.Fatalf("batch load: %v", err)
	}
	if len(byResource[postA]) != 1 || len(byResource[postB]) != 1 {
		t.Fatalf("expected A and B to have one category each, got %d and %d", len(byResource[postA]), len(byResource[postB]))
	}
	if len(byResource[postC]) != 0 {
		t.Fatalf("expected C to have no categories, got %d", len(byResource[postC]))
	}

	single, err := svc.GetCategoriesForResource(ctx, "post", postA)
	if err != nil {
		t.Fatalf("single load: %v", err)
	}
	if len(single) != 1 || single[0].ID != cat {
		t.Fatalf("single-resource load mismatch: %+v", single)
	}
}

func TestService_TagsBatchLoad(t *testing.T) {
	db, svc := setupServiceTestDB(t)
	ctx := context.Background()

	tag := uuid.New()
	if _, err := db.Exec(ctx, "INSERT INTO tags (id, name, slug, created_at) VALUES (?, ?, ?, ?)", tag, "go", "go", nowStamp()); err != nil {
		t.Fatalf("insert tag: %v", err)
	}

	post := uuid.New()
	if err := svc.AttachTag(ctx, tag, "post", post); err != nil {
		t.Fatalf("attach tag: %v", err)
	}

	byResource, err := svc.GetTagsForResources(ctx, "post", []uuid.UUID{post})
	if err != nil {
		t.Fatalf("batch tag load: %v", err)
	}
	if len(byResource[post]) != 1 || byResource[post][0].ID != tag {
		t.Fatalf("tag batch mismatch: %+v", byResource)
	}
}

func TestService_CategoryTreeCache(t *testing.T) {
	db, svc := setupServiceTestDB(t)
	ctx := context.Background()

	root := uuid.New()
	insertCategory(t, db, root, nil, "root")

	first, err := svc.GetAllCategories(ctx)
	if err != nil {
		t.Fatalf("first load: %v", err)
	}
	if len(first) != 1 {
		t.Fatalf("expected 1 category, got %d", len(first))
	}

	insertCategory(t, db, uuid.New(), &root, "child")

	cached, err := svc.GetAllCategories(ctx)
	if err != nil {
		t.Fatalf("cached load: %v", err)
	}
	if len(cached) != 1 {
		t.Fatalf("expected cache to hide the uninvalidated insert, got %d", len(cached))
	}

	svc.InvalidateCategoryCache()
	fresh, err := svc.GetAllCategories(ctx)
	if err != nil {
		t.Fatalf("fresh load: %v", err)
	}
	if len(fresh) != 2 {
		t.Fatalf("expected 2 categories after invalidation, got %d", len(fresh))
	}

	tree := svc.BuildCategoryTree(fresh)
	if len(tree) != 1 || len(tree[0].Children) != 1 {
		t.Fatalf("expected one root with one child, got %+v", tree)
	}
}

func TestService_CategoryDepthSingleQuery(t *testing.T) {
	db, svc := setupServiceTestDB(t)
	ctx := context.Background()

	a, b, c := uuid.New(), uuid.New(), uuid.New()
	insertCategory(t, db, a, nil, "a")
	insertCategory(t, db, b, &a, "b")
	insertCategory(t, db, c, &b, "c")

	depth, err := svc.GetCategoryDepth(ctx, c)
	if err != nil {
		t.Fatalf("depth: %v", err)
	}
	if depth != 2 {
		t.Fatalf("expected depth 2, got %d", depth)
	}

	if _, err := svc.GetCategoryDepth(ctx, uuid.New()); err == nil {
		t.Fatal("expected error for non-existent category")
	}
}

func TestService_ResourceIDsBySlug(t *testing.T) {
	db, svc := setupServiceTestDB(t)
	ctx := context.Background()

	cat := uuid.New()
	insertCategory(t, db, cat, nil, "featured")
	post := uuid.New()
	if err := svc.AttachCategory(ctx, cat, "post", post); err != nil {
		t.Fatalf("attach: %v", err)
	}

	ids, err := svc.GetResourceIDsByCategorySlug(ctx, "post", "featured")
	if err != nil {
		t.Fatalf("by slug: %v", err)
	}
	if len(ids) != 1 || ids[0] != post {
		t.Fatalf("expected post id, got %+v", ids)
	}
}
