package handlers

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/mattn/go-sqlite3"
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

	return c.Status(http.StatusCreated).JSON(fiber.Map{"id": res.(int64)})
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

func isUniqueViolation(err error) bool {
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) {
		return sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique
	}
	return false
}
