# gorest-taxonomy

[![CI](https://github.com/nicolasbonnici/gorest-taxonomy/actions/workflows/ci.yml/badge.svg)](https://github.com/nicolasbonnici/gorest-taxonomy/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/nicolasbonnici/gorest-taxonomy)](https://goreportcard.com/report/github.com/nicolasbonnici/gorest-taxonomy)
[![Go Version](https://img.shields.io/badge/go-1.26-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Polymorphic **categories** (hierarchical) and **tags** (flat) for any GoREST resource type. Attach taxonomy to any entity via `resource` + `resource_id` — no schema changes needed on the target table.

## Schema

| Table | Purpose |
|---|---|
| `categories` | Self-referential tree (`parent_id = null` → root) |
| `category_resources` | Pivot: category ↔ any resource |
| `tags` | Flat tag list |
| `tag_resources` | Pivot: tag ↔ any resource |

Supports PostgreSQL, MySQL, SQLite.

## Installation

```bash
go get github.com/nicolasbonnici/gorest-taxonomy
```

Register in `gorest.yaml`:

```yaml
plugins:
  - name: taxonomy
    config:
      allowed_types: [post, product, article]
      max_depth: 5          # max category nesting levels (default: 5)
      pagination_limit: 25
      max_pagination_limit: 100
```

## API

### Categories

```
POST   /categories                                    create (slug auto-generated from name)
GET    /categories                                    list (paginated, filterable)
GET    /categories/tree                               full tree
GET    /categories/:id                               get by id
PUT    /categories/:id                               full update
DELETE /categories/:id                               delete (children become root)
POST   /categories/:id/resources                     attach { resource, resource_id }
DELETE /categories/:id/resources/:resource/:rid      detach
GET    /:resource/:resource_id/categories            list categories for a resource
```

### Tags

```
POST   /tags                                         create
GET    /tags                                         list (paginated, filterable)
GET    /tags/:id                                     get by id
PUT    /tags/:id                                     full update
DELETE /tags/:id                                     delete
POST   /tags/:id/resources                           attach { resource, resource_id }
DELETE /tags/:id/resources/:resource/:rid            detach
GET    /:resource/:resource_id/tags                  list tags for a resource
```

## Examples

```bash
# Create a root category
curl -X POST /categories -d '{"name":"Technology"}'

# Create a child category
curl -X POST /categories -d '{"name":"Go","parent_id":"<uuid>"}'

# Get full tree
curl /categories/tree

# Attach a post to a category
curl -X POST /categories/<uuid>/resources \
  -d '{"resource":"post","resource_id":"<uuid>"}'

# List all categories for a post
curl /post/<uuid>/categories

# Create and attach a tag
curl -X POST /tags -d '{"name":"open-source"}'
curl -X POST /tags/<uuid>/resources \
  -d '{"resource":"post","resource_id":"<uuid>"}'
```

## Development

```bash
make install    # deps + dev tools + git hooks
make test       # run tests
make audit      # gofmt, vet, staticcheck, errcheck, gocyclo
make lint       # golangci-lint
```

## License

MIT
