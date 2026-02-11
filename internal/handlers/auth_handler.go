package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"modernc.org/sqlite"
)

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var req authRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	res, err := h.runInPool(c.UserContext(), func() (any, error) {
		return h.auth.Register(req.Email, req.Password)
	})
	if err != nil {
		if isUniqueViolation(err) {
			return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "email already registered"})
		}
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	id, ok := res.(int64)
	if !ok {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "unexpected response type"})
	}
	return c.Status(http.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req authRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	res, err := h.runInPool(c.UserContext(), func() (any, error) {
		token, user, err := h.auth.Login(req.Email, req.Password)
		if err != nil {
			return nil, err
		}
		return fiber.Map{
			"token": token,
			"user": fiber.Map{
				"id":    user.ID,
				"email": user.Email,
			},
		}, nil
	})
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	return c.JSON(res)
}

// Detect UNIQUE constraint violations with modernc.org/sqlite.
// First try typed match; then fall back to stable message fragment.

func isUniqueViolation(err error) bool {
	var se *sqlite.Error
	if errors.As(err, &se) {
		// modernc exposes extended codes via Code()
		const sqliteConstraintUnique = 2067 // SQLITE_CONSTRAINT_UNIQUE
		if int(se.Code()) == sqliteConstraintUnique {
			return true
		}
	}

	// Fallback (safe backup)
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique constraint failed")
}
