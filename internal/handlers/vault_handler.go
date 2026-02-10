package handlers

import (
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"vault/internal/models"
)

type vaultRequest struct {
	Title    string `json:"title"`
	Username string `json:"username"`
	Password string `json:"password"`
	URL      string `json:"url"`
	Category string `json:"category"`
	Notes    string `json:"notes"`
}

func (h *Handler) ListEntries(c *fiber.Ctx) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	res, err := h.runInPool(c.UserContext(), func() (any, error) {
		return h.vault.List(userID)
	})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "could not load entries"})
	}

	return c.JSON(res)
}

func (h *Handler) GetEntry(c *fiber.Ctx) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	res, err := h.runInPool(c.UserContext(), func() (any, error) {
		return h.vault.Get(userID, id)
	})
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "entry not found"})
	}

	return c.JSON(res)
}

func (h *Handler) CreateEntry(c *fiber.Ctx) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req vaultRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	entry := models.VaultEntry{
		Title:    req.Title,
		Username: req.Username,
		Password: req.Password,
		URL:      req.URL,
		Category: req.Category,
		Notes:    req.Notes,
	}

	res, err := h.runInPool(c.UserContext(), func() (any, error) {
		return h.vault.Create(userID, entry)
	})
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"id": res.(int64)})
}

func (h *Handler) UpdateEntry(c *fiber.Ctx) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var req vaultRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	entry := models.VaultEntry{
		Title:    req.Title,
		Username: req.Username,
		Password: req.Password,
		URL:      req.URL,
		Category: req.Category,
		Notes:    req.Notes,
	}

	_, err = h.runInPool(c.UserContext(), func() (any, error) {
		return nil, h.vault.Update(userID, id, entry)
	})
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(http.StatusNoContent)
}

func (h *Handler) DeleteEntry(c *fiber.Ctx) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	_, err = h.runInPool(c.UserContext(), func() (any, error) {
		return nil, h.vault.Delete(userID, id)
	})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "could not delete"})
	}

	return c.SendStatus(http.StatusNoContent)
}

func (h *Handler) SearchEntries(c *fiber.Ctx) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	query := c.Query("q", "")

	res, err := h.runInPool(c.UserContext(), func() (any, error) {
		return h.vault.Search(c.UserContext(), userID, query)
	})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "search failed", "details": err.Error()})
	}

	return c.JSON(res)
}
