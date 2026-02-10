package handlers

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func userIDFromToken(c *fiber.Ctx) (int64, error) {
	user := c.Locals("user")
	token, ok := user.(*jwt.Token)
	if !ok || token == nil {
		return 0, errors.New("missing token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims")
	}

	sub, ok := claims["sub"]
	if !ok {
		return 0, errors.New("missing sub")
	}

	switch value := sub.(type) {
	case float64:
		return int64(value), nil
	case string:
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid sub: %w", err)
		}
		return id, nil
	default:
		return 0, errors.New("invalid sub type")
	}
}
