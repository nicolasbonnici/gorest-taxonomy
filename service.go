package taxonomy

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/nicolasbonnici/gorest/database"
)

type TaxonomyService struct {
	db     database.Database
	config *Config
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

func (s *TaxonomyService) GetAllCategories(ctx context.Context) ([]Category, error) {
	rows, err := s.db.Query(ctx, "SELECT id, parent_id, name, slug, description, created_at, updated_at FROM categories ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.ParentID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (s *TaxonomyService) GetCategoriesForResource(ctx context.Context, resource string, resourceID uuid.UUID) ([]Category, error) {
	d := s.db.Dialect()
	sql := "SELECT c.id, c.parent_id, c.name, c.slug, c.description, c.created_at, c.updated_at " +
		"FROM categories c " +
		"INNER JOIN category_resources cr ON cr.category_id = c.id " +
		"WHERE cr.resource = " + d.Placeholder(1) + " AND cr.resource_id = " + d.Placeholder(2)

	rows, err := s.db.Query(ctx, sql, resource, resourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.ParentID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (s *TaxonomyService) GetTagsForResource(ctx context.Context, resource string, resourceID uuid.UUID) ([]Tag, error) {
	d := s.db.Dialect()
	sql := "SELECT t.id, t.name, t.slug, t.created_at, t.updated_at " +
		"FROM tags t " +
		"INNER JOIN tag_resources tr ON tr.tag_id = t.id " +
		"WHERE tr.resource = " + d.Placeholder(1) + " AND tr.resource_id = " + d.Placeholder(2)

	rows, err := s.db.Query(ctx, sql, resource, resourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (s *TaxonomyService) GetCategoryDepth(ctx context.Context, id uuid.UUID) (int, error) {
	d := s.db.Dialect()
	depth := 0
	currentID := id
	for depth <= s.config.MaxDepth {
		var parentID *uuid.UUID
		if err := s.db.QueryRow(ctx, "SELECT parent_id FROM categories WHERE id = "+d.Placeholder(1), currentID).Scan(&parentID); err != nil {
			return 0, err
		}
		if parentID == nil {
			return depth, nil
		}
		depth++
		currentID = *parentID
	}
	return depth, nil
}

func (s *TaxonomyService) AttachCategory(ctx context.Context, categoryID uuid.UUID, resource string, resourceID uuid.UUID) error {
	d := s.db.Dialect()
	sql := "INSERT INTO category_resources (id, category_id, resource, resource_id, created_at) VALUES (" +
		d.Placeholder(1) + ", " + d.Placeholder(2) + ", " + d.Placeholder(3) + ", " + d.Placeholder(4) + ", " + d.Placeholder(5) + ")"
	_, err := s.db.Exec(ctx, sql, uuid.New(), categoryID, resource, resourceID, time.Now())
	return err
}

func (s *TaxonomyService) DetachCategory(ctx context.Context, categoryID uuid.UUID, resource string, resourceID uuid.UUID) error {
	d := s.db.Dialect()
	sql := "DELETE FROM category_resources WHERE category_id = " + d.Placeholder(1) +
		" AND resource = " + d.Placeholder(2) + " AND resource_id = " + d.Placeholder(3)
	_, err := s.db.Exec(ctx, sql, categoryID, resource, resourceID)
	return err
}

func (s *TaxonomyService) AttachTag(ctx context.Context, tagID uuid.UUID, resource string, resourceID uuid.UUID) error {
	d := s.db.Dialect()
	sql := "INSERT INTO tag_resources (id, tag_id, resource, resource_id, created_at) VALUES (" +
		d.Placeholder(1) + ", " + d.Placeholder(2) + ", " + d.Placeholder(3) + ", " + d.Placeholder(4) + ", " + d.Placeholder(5) + ")"
	_, err := s.db.Exec(ctx, sql, uuid.New(), tagID, resource, resourceID, time.Now())
	return err
}

func (s *TaxonomyService) DetachTag(ctx context.Context, tagID uuid.UUID, resource string, resourceID uuid.UUID) error {
	d := s.db.Dialect()
	sql := "DELETE FROM tag_resources WHERE tag_id = " + d.Placeholder(1) +
		" AND resource = " + d.Placeholder(2) + " AND resource_id = " + d.Placeholder(3)
	_, err := s.db.Exec(ctx, sql, tagID, resource, resourceID)
	return err
}
