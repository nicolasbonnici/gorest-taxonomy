package taxonomy

import (
	"testing"
)

func TestNewPlugin(t *testing.T) {
	p := NewPlugin()
	if p == nil {
		t.Fatal("NewPlugin() returned nil")
	}
	if p.Name() != "taxonomy" {
		t.Errorf("Name() = %v, want 'taxonomy'", p.Name())
	}
}

func TestTaxonomyPlugin_Initialize(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name:    "empty config uses defaults",
			config:  map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "valid config",
			config: map[string]interface{}{
				"allowed_types":        []interface{}{"post", "article"},
				"max_depth":            5,
				"pagination_limit":     25,
				"max_pagination_limit": 100,
			},
			wantErr: false,
		},
		{
			name: "invalid allowed_types",
			config: map[string]interface{}{
				"allowed_types": []interface{}{"post", ""},
			},
			wantErr: true,
		},
		{
			name: "max_depth out of range",
			config: map[string]interface{}{
				"allowed_types": []interface{}{"post"},
				"max_depth":     25,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &TaxonomyPlugin{}
			err := p.Initialize(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaxonomyPlugin_MigrationDependencies(t *testing.T) {
	p := &TaxonomyPlugin{}
	if len(p.MigrationDependencies()) != 0 {
		t.Error("MigrationDependencies() should return empty slice")
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello-world"},
		{"Blog & News", "blog-news"},
		{"  spaces  ", "spaces"},
		{"UPPERCASE", "uppercase"},
		{"special!@#chars", "special-chars"},
		{"multiple---hyphens", "multiple-hyphens"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := slugify(tt.input); got != tt.want {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
