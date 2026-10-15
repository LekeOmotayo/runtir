//go:build tools

package tools

import (
	_ "github.com/gin-gonic/gin"
	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/google/uuid"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/joho/godotenv"
	_ "github.com/pressly/goose/v3"
	_ "golang.org/x/crypto/bcrypt"
)
