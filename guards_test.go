package taxonomy

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/nicolasbonnici/gorest/rbac"
)

// Category and tag mutations carried no guard before v0.7, so anyone could
// create, edit or delete taxonomy terms anonymously. These lock that shut.

const guardUserID = "5d8a1d4e-0f0a-4f27-9f33-9b6e5f0d2c11"

func newGuardApp(t *testing.T, auth fiber.Handler) *fiber.App {
	t.Helper()
	db, svc := setupServiceTestDB(t)
	t.Cleanup(func() { _ = db.Close() })

	cfg := DefaultConfig()
	cfg.Database = db
	cfg.AuthMiddleware = auth

	app := fiber.New()
	RegisterRoutesWithService(app, db, &cfg, svc)
	return app
}

// stubIdentity stands in for the host's auth middleware. It seeds roles
// directly because RoleLoader reads them from tables this fixture omits, and it
// leaves an existing context alone when it finds none.
func stubIdentity(userID string, roles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		if userID != "" {
			c.Locals("user_id", userID)
			c.SetContext(rbac.WithRoles(c.Context(), roles))
		}
		return c.Next()
	}
}

func jsonReq(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestTaxonomyMutationsRejectAnonymous(t *testing.T) {
	app := newGuardApp(t, nil)

	cases := []struct {
		name string
		req  *http.Request
	}{
		{"create category", jsonReq(http.MethodPost, "/categories", `{"name":"x","slug":"x"}`)},
		{"update category", jsonReq(http.MethodPut, "/categories/"+guardUserID, `{"name":"x"}`)},
		{"delete category", httptest.NewRequest(http.MethodDelete, "/categories/"+guardUserID, nil)},
		{"create tag", jsonReq(http.MethodPost, "/tags", `{"name":"x","slug":"x"}`)},
		{"delete tag", httptest.NewRequest(http.MethodDelete, "/tags/"+guardUserID, nil)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := app.Test(tc.req)
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("status = %d, want 401", resp.StatusCode)
			}
		})
	}
}

func TestTaxonomyMutationsRejectReader(t *testing.T) {
	app := newGuardApp(t, stubIdentity(guardUserID, "reader"))

	resp, err := app.Test(jsonReq(http.MethodPost, "/categories", `{"name":"x","slug":"x"}`))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d, want 403", resp.StatusCode)
	}
}

func TestTaxonomyWriterMayMutate(t *testing.T) {
	app := newGuardApp(t, stubIdentity(guardUserID, "writer"))

	resp, err := app.Test(jsonReq(http.MethodPost, "/categories", `{"name":"guard","slug":"guard"}`))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		t.Errorf("writer was denied: status = %d", resp.StatusCode)
	}
}

func TestTaxonomyReadsStayPublic(t *testing.T) {
	app := newGuardApp(t, nil)

	for _, path := range []string{"/categories", "/tags"} {
		resp, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET %s status = %d, want 200", path, resp.StatusCode)
		}
	}
}
