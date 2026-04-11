package config

import "github.com/joho/godotenv"

// LoadDotenv loads `.env` from the current directory, then from the parent
// directory (repo root when cwd is backend/). Later files only set variables
// that are not already defined (including from the process environment).
func LoadDotenv() {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
}
