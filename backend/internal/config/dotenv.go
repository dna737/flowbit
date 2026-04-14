package config

import (
	"path/filepath"

	"github.com/joho/godotenv"
)

// dotenvDir is the absolute directory of the first .env file successfully
// loaded. Used to resolve relative file paths in the config (e.g. cert files).
var dotenvDir string

// LoadDotenv loads `.env` from the current directory, then from parent
// directories up to two levels (repo root when cwd is backend/ or
// backend/integration/). Later files only set variables not already defined.
// The first successfully loaded file's directory is used to anchor relative
// paths in the config.
func LoadDotenv() {
	for _, rel := range []string{"../.env", "../../.env"} {
		if err := godotenv.Load(rel); err == nil && dotenvDir == "" {
			if abs, err := filepath.Abs(filepath.Dir(rel)); err == nil {
				dotenvDir = abs
			}
		}
	}
}
