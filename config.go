package taxonomy

import (
	"errors"
	"fmt"

	"github.com/nicolasbonnici/gorest/database"
)

type Config struct {
	Database           database.Database
	AllowedTypes       []string `json:"allowed_types" yaml:"allowed_types"`
	MaxDepth           int      `json:"max_depth" yaml:"max_depth"`
	PaginationLimit    int      `json:"pagination_limit" yaml:"pagination_limit"`
	MaxPaginationLimit int      `json:"max_pagination_limit" yaml:"max_pagination_limit"`
}

func DefaultConfig() Config {
	return Config{
		AllowedTypes:       []string{"post"},
		MaxDepth:           5,
		PaginationLimit:    25,
		MaxPaginationLimit: 100,
	}
}

func (c *Config) Validate() error {
	if err := c.validateAllowedTypes(); err != nil {
		return err
	}

	c.applyDefaults()

	if c.MaxDepth < 1 || c.MaxDepth > 20 {
		return errors.New("max_depth must be between 1 and 20")
	}

	return nil
}

func (c *Config) validateAllowedTypes() error {
	if len(c.AllowedTypes) == 0 {
		return errors.New("allowed_types cannot be empty")
	}

	seen := make(map[string]bool)
	for _, t := range c.AllowedTypes {
		if t == "" {
			return errors.New("allowed_types cannot contain empty strings")
		}
		if seen[t] {
			return fmt.Errorf("duplicate type in allowed_types: %s", t)
		}
		seen[t] = true
	}

	return nil
}

func (c *Config) applyDefaults() {
	if c.PaginationLimit <= 0 {
		c.PaginationLimit = 25
	}
	if c.MaxPaginationLimit <= 0 {
		c.MaxPaginationLimit = 100
	}
	if c.MaxDepth <= 0 {
		c.MaxDepth = 5
	}
}

func (c *Config) IsAllowedType(t string) bool {
	for _, allowed := range c.AllowedTypes {
		if allowed == t {
			return true
		}
	}
	return false
}
