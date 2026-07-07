package taxonomy

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/query"
)

type TaxonomyService struct {
	db     database.Database
	config *Config

	// The full category set is memoized because the tree endpoint and depth
	// checks re-read every category on each call while writes are rare. The
	// cache is dropped wholesale on any category mutation (see the category
	// handlers) rather than surgically patched: the set is small and bounded
	// by MaxDepth, so a full reload is cheaper than tracking deltas.
	categoryCacheMu sync.RWMutex
	categoryCache   []Category
	categoryCached  bool
}

func NewTaxonomyService(db database.Database, config *Config) *TaxonomyService {
	return &TaxonomyService{db: db, config: config}
}

func (s *TaxonomyService) BuildCategoryTree(categories []Category) []*CategoryTreeNode {
	nodes := make(map[uuid.UUID]*CategoryTreeNode, len(categories))
	for i := range categories {
		c := &categories[i]
		nodes[c.ID] = &CategoryTreeNode{
			ID:          c.ID,
			ParentID:    c.ParentID,
			Name:        c.Name,
			Slug:        c.Slug,
			Description: c.Description,
			CreatedAt:   c.CreatedAt,
			UpdatedAt:   c.UpdatedAt,
		}
	}

	var roots []*CategoryTreeNode
	for _, node := range nodes {
		if node.ParentID == nil {
			roots = append(roots, node)
		} else if parent, ok := nodes[*node.ParentID]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			roots = append(roots, node)
		}
	}

	sort.Slice(roots, func(i, j int) bool { return roots[i].Name < roots[j].Name })
	for _, node := range nodes {
		sort.Slice(node.Children, func(i, j int) bool { return node.Children[i].Name < node.Children[j].Name })
	}

	return roots
}

// InvalidateCategoryCache drops the memoized category set. It must be called
// after any committed create/update/delete of a category so the next read
// reloads fresh data.
func (s *TaxonomyService) InvalidateCategoryCache() {
	s.categoryCacheMu.Lock()
	s.categoryCache = nil
	s.categoryCached = false
	s.categoryCacheMu.Unlock()
}

func (s *TaxonomyService) GetAllCategories(ctx context.Context) ([]Category, error) {
	s.categoryCacheMu.RLock()
	if s.categoryCached {
		out := cloneCategories(s.categoryCache)
		s.categoryCacheMu.RUnlock()
		return out, nil
	}
	s.categoryCacheMu.RUnlock()

	categories, err := s.loadAllCategories(ctx)
	if err != nil {
		return nil, err
	}

	s.categoryCacheMu.Lock()
	s.categoryCache = categories
	s.categoryCached = true
	s.categoryCacheMu.Unlock()

	// Callers receive an independent copy so the cached slice can never be
	// mutated from the outside.
	return cloneCategories(categories), nil
}

func (s *TaxonomyService) loadAllCategories(ctx context.Context) ([]Category, error) {
	sql, args, err := query.New(s.db.Dialect()).
		Select("id", "parent_id", "name", "slug", "description", "created_at", "updated_at").
		From("categories").
		OrderBy("name", query.ASC).
		Build()
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var c Category
		var created, updated portableTime
		if err := rows.Scan(&c.ID, &c.ParentID, &c.Name, &c.Slug, &c.Description, &created, &updated); err != nil {
			return nil, err
		}
		c.CreatedAt = created.Time
		c.UpdatedAt = updated.Ptr()
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

// GetCategoriesForResource returns the categories attached to a single resource.
// It delegates to the batch loader so both paths share one query shape.
func (s *TaxonomyService) GetCategoriesForResource(ctx context.Context, resource string, resourceID uuid.UUID) ([]Category, error) {
	byResource, err := s.GetCategoriesForResources(ctx, resource, []uuid.UUID{resourceID})
	if err != nil {
		return nil, err
	}
	return byResource[resourceID], nil
}

// GetCategoriesForResources loads categories for a batch of resources of the
// same type in a single query, keyed by resource id. This is the N+1 escape
// hatch for consumers that render categories across a page of resources: one
// round trip instead of one per resource.
func (s *TaxonomyService) GetCategoriesForResources(ctx context.Context, resource string, resourceIDs []uuid.UUID) (map[uuid.UUID][]Category, error) {
	result := make(map[uuid.UUID][]Category, len(resourceIDs))
	if len(resourceIDs) == 0 {
		return result, nil
	}

	sql, args, err := query.New(s.db.Dialect()).
		Select("cr.resource_id", "c.id", "c.parent_id", "c.name", "c.slug", "c.description", "c.created_at", "c.updated_at").
		From("categories").As("c").
		JoinAs("category_resources", "cr", query.ColEq("cr.category_id", "c.id")).
		Where(query.Eq("cr.resource", resource)).
		Where(query.In("cr.resource_id", toAnySlice(resourceIDs)...)).
		Build()
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var resourceID uuid.UUID
		var c Category
		var created, updated portableTime
		if err := rows.Scan(&resourceID, &c.ID, &c.ParentID, &c.Name, &c.Slug, &c.Description, &created, &updated); err != nil {
			return nil, err
		}
		c.CreatedAt = created.Time
		c.UpdatedAt = updated.Ptr()
		result[resourceID] = append(result[resourceID], c)
	}
	return result, rows.Err()
}

// GetTagsForResource returns the tags attached to a single resource, delegating
// to the batch loader.
func (s *TaxonomyService) GetTagsForResource(ctx context.Context, resource string, resourceID uuid.UUID) ([]Tag, error) {
	byResource, err := s.GetTagsForResources(ctx, resource, []uuid.UUID{resourceID})
	if err != nil {
		return nil, err
	}
	return byResource[resourceID], nil
}

// GetTagsForResources loads tags for a batch of resources of the same type in a
// single query, keyed by resource id.
func (s *TaxonomyService) GetTagsForResources(ctx context.Context, resource string, resourceIDs []uuid.UUID) (map[uuid.UUID][]Tag, error) {
	result := make(map[uuid.UUID][]Tag, len(resourceIDs))
	if len(resourceIDs) == 0 {
		return result, nil
	}

	sql, args, err := query.New(s.db.Dialect()).
		Select("tr.resource_id", "t.id", "t.name", "t.slug", "t.created_at", "t.updated_at").
		From("tags").As("t").
		JoinAs("tag_resources", "tr", query.ColEq("tr.tag_id", "t.id")).
		Where(query.Eq("tr.resource", resource)).
		Where(query.In("tr.resource_id", toAnySlice(resourceIDs)...)).
		Build()
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var resourceID uuid.UUID
		var t Tag
		var created, updated portableTime
		if err := rows.Scan(&resourceID, &t.ID, &t.Name, &t.Slug, &created, &updated); err != nil {
			return nil, err
		}
		t.CreatedAt = created.Time
		t.UpdatedAt = updated.Ptr()
		result[resourceID] = append(result[resourceID], t)
	}
	return result, rows.Err()
}

// GetCategoryDepth returns how many ancestors a category has. It loads the
// id→parent map in one query and walks it in memory instead of issuing a query
// per level, which turned the depth check into up to MaxDepth round trips. The
// map is read straight from the database (not the cache) so a parent created
// concurrently is always visible to the depth guard.
func (s *TaxonomyService) GetCategoryDepth(ctx context.Context, id uuid.UUID) (int, error) {
	parents, err := s.loadCategoryParents(ctx)
	if err != nil {
		return 0, err
	}
	if _, exists := parents[id]; !exists {
		return 0, fmt.Errorf("category %s does not exist", id)
	}

	depth := 0
	current := id
	for depth <= s.config.MaxDepth {
		parent := parents[current]
		if parent == nil {
			return depth, nil
		}
		depth++
		current = *parent
	}
	return depth, nil
}

func (s *TaxonomyService) loadCategoryParents(ctx context.Context) (map[uuid.UUID]*uuid.UUID, error) {
	sql, args, err := query.New(s.db.Dialect()).
		Select("id", "parent_id").
		From("categories").
		Build()
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	parents := make(map[uuid.UUID]*uuid.UUID)
	for rows.Next() {
		var id uuid.UUID
		var parentID *uuid.UUID
		if err := rows.Scan(&id, &parentID); err != nil {
			return nil, err
		}
		parents[id] = parentID
	}
	return parents, rows.Err()
}

func (s *TaxonomyService) AttachCategory(ctx context.Context, categoryID uuid.UUID, resource string, resourceID uuid.UUID) error {
	sql, args, err := query.New(s.db.Dialect()).
		Insert("category_resources").
		Columns("id", "category_id", "resource", "resource_id", "created_at").
		Values(uuid.New(), categoryID, resource, resourceID, time.Now()).
		Build()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, sql, args...)
	return err
}

func (s *TaxonomyService) DetachCategory(ctx context.Context, categoryID uuid.UUID, resource string, resourceID uuid.UUID) error {
	sql, args, err := query.New(s.db.Dialect()).
		Delete("category_resources").
		Where(query.Eq("category_id", categoryID)).
		Where(query.Eq("resource", resource)).
		Where(query.Eq("resource_id", resourceID)).
		Build()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, sql, args...)
	return err
}

func (s *TaxonomyService) AttachTag(ctx context.Context, tagID uuid.UUID, resource string, resourceID uuid.UUID) error {
	sql, args, err := query.New(s.db.Dialect()).
		Insert("tag_resources").
		Columns("id", "tag_id", "resource", "resource_id", "created_at").
		Values(uuid.New(), tagID, resource, resourceID, time.Now()).
		Build()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, sql, args...)
	return err
}

func (s *TaxonomyService) DetachTag(ctx context.Context, tagID uuid.UUID, resource string, resourceID uuid.UUID) error {
	sql, args, err := query.New(s.db.Dialect()).
		Delete("tag_resources").
		Where(query.Eq("tag_id", tagID)).
		Where(query.Eq("resource", resource)).
		Where(query.Eq("resource_id", resourceID)).
		Build()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, sql, args...)
	return err
}

func (s *TaxonomyService) GetResourceIDsByCategorySlug(ctx context.Context, resource, slug string) ([]uuid.UUID, error) {
	sql, args, err := query.New(s.db.Dialect()).
		Select("cr.resource_id").
		From("category_resources").As("cr").
		JoinAs("categories", "c", query.ColEq("c.id", "cr.category_id")).
		Where(query.Eq("cr.resource", resource)).
		Where(query.Eq("c.slug", slug)).
		Build()
	if err != nil {
		return nil, err
	}
	return s.scanResourceIDs(ctx, sql, args)
}

func (s *TaxonomyService) GetResourceIDsByTagSlug(ctx context.Context, resource, slug string) ([]uuid.UUID, error) {
	sql, args, err := query.New(s.db.Dialect()).
		Select("tr.resource_id").
		From("tag_resources").As("tr").
		JoinAs("tags", "t", query.ColEq("t.id", "tr.tag_id")).
		Where(query.Eq("tr.resource", resource)).
		Where(query.Eq("t.slug", slug)).
		Build()
	if err != nil {
		return nil, err
	}
	return s.scanResourceIDs(ctx, sql, args)
}

func (s *TaxonomyService) scanResourceIDs(ctx context.Context, sql string, args []any) ([]uuid.UUID, error) {
	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func cloneCategories(in []Category) []Category {
	if in == nil {
		return nil
	}
	out := make([]Category, len(in))
	copy(out, in)
	return out
}

func toAnySlice(ids []uuid.UUID) []any {
	out := make([]any, len(ids))
	for i, id := range ids {
		out[i] = id
	}
	return out
}

// portableTime scans a timestamp column across dialects: Postgres returns a
// time.Time, while SQLite stores it as TEXT and hands back a string. Scanning
// straight into time.Time would fail on SQLite, so column reads go through this.
type portableTime struct {
	Time  time.Time
	Valid bool
}

var timestampLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02 15:04:05.999999999-07:00",
	"2006-01-02 15:04:05.999999999",
	"2006-01-02 15:04:05",
}

func (p *portableTime) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		p.Time, p.Valid = time.Time{}, false
	case time.Time:
		p.Time, p.Valid = v, true
	case []byte:
		return p.Scan(string(v))
	case string:
		for _, layout := range timestampLayouts {
			if parsed, err := time.Parse(layout, v); err == nil {
				p.Time, p.Valid = parsed, true
				return nil
			}
		}
		return fmt.Errorf("cannot parse timestamp %q", v)
	default:
		return fmt.Errorf("unsupported timestamp type %T", value)
	}
	return nil
}

func (p portableTime) Ptr() *time.Time {
	if !p.Valid {
		return nil
	}
	t := p.Time
	return &t
}
