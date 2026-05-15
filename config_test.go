package taxonomy

import (
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				AllowedTypes:       []string{"post", "article"},
				MaxDepth:           5,
				PaginationLimit:    25,
				MaxPaginationLimit: 100,
			},
			wantErr: false,
		},
		{
			name: "empty allowed_types",
			config: Config{
				AllowedTypes: []string{},
			},
			wantErr: true,
			errMsg:  "allowed_types cannot be empty",
		},
		{
			name: "allowed_types with empty string",
			config: Config{
				AllowedTypes: []string{"post", ""},
			},
			wantErr: true,
			errMsg:  "allowed_types cannot contain empty strings",
		},
		{
			name: "duplicate allowed_types",
			config: Config{
				AllowedTypes: []string{"post", "article", "post"},
			},
			wantErr: true,
			errMsg:  "duplicate type in allowed_types: post",
		},
		{
			name: "max_depth too large",
			config: Config{
				AllowedTypes: []string{"post"},
				MaxDepth:     21,
			},
			wantErr: true,
			errMsg:  "max_depth must be between 1 and 20",
		},
		{
			name: "max_depth too small",
			config: Config{
				AllowedTypes: []string{"post"},
				MaxDepth:     0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestConfig_IsAllowedType(t *testing.T) {
	config := Config{
		AllowedTypes: []string{"post", "product", "article"},
	}

	tests := []struct {
		typeName string
		want     bool
	}{
		{"post", true},
		{"product", true},
		{"article", true},
		{"comment", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			if got := config.IsAllowedType(tt.typeName); got != tt.want {
				t.Errorf("IsAllowedType(%q) = %v, want %v", tt.typeName, got, tt.want)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if len(config.AllowedTypes) == 0 {
		t.Error("DefaultConfig() AllowedTypes should not be empty")
	}
	if config.MaxDepth != 5 {
		t.Errorf("DefaultConfig() MaxDepth = %d, want 5", config.MaxDepth)
	}
	if config.PaginationLimit != 25 {
		t.Errorf("DefaultConfig() PaginationLimit = %d, want 25", config.PaginationLimit)
	}
	if config.MaxPaginationLimit != 100 {
		t.Errorf("DefaultConfig() MaxPaginationLimit = %d, want 100", config.MaxPaginationLimit)
	}
}
