package taxonomy

import (
	"testing"
)

func TestCategory_TableName(t *testing.T) {
	var m Category
	if m.TableName() != "categories" {
		t.Errorf("Category.TableName() = %v, want 'categories'", m.TableName())
	}
}

func TestCategoryResource_TableName(t *testing.T) {
	var m CategoryResource
	if m.TableName() != "category_resources" {
		t.Errorf("CategoryResource.TableName() = %v, want 'category_resources'", m.TableName())
	}
}

func TestTag_TableName(t *testing.T) {
	var m Tag
	if m.TableName() != "tags" {
		t.Errorf("Tag.TableName() = %v, want 'tags'", m.TableName())
	}
}

func TestTagResource_TableName(t *testing.T) {
	var m TagResource
	if m.TableName() != "tag_resources" {
		t.Errorf("TagResource.TableName() = %v, want 'tag_resources'", m.TableName())
	}
}
