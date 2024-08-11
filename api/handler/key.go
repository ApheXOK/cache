package handler

import (
	"cache/api/presenter"
	"cache/database"
	"cache/model"
	"database/sql"
	"errors"
	"github.com/gofiber/fiber/v2"
	"strings"
)

type GetKeysReq struct {
	Kids string `json:"kids"`
}

func GetKeys(db *database.BunDB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		req := new(GetKeysReq)
		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request",
			})
		}
		found, err := db.GetKeys(c.Context(), strings.Split(req.Kids, ","))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.KeyErrorResponse(err))
		}
		if len(found) == 0 {
			return c.Status(fiber.StatusNotFound).JSON(presenter.KeyErrorResponse(errors.New("key not found")))
		}
		return c.Status(200).JSON(presenter.KeyListSuccessResponse(found))
	}
}

func GetKey(db *database.BunDB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		kid := c.Params("kid")
		found, err := db.GetKey(c.Context(), kid)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.KeyErrorResponse(err))
		}
		if found == nil {
			return c.Status(fiber.StatusNotFound).JSON(presenter.KeyErrorResponse(errors.New("key not found")))
		}
		return c.Status(200).JSON(presenter.KeySuccessResponse(found))
	}
}

func SaveKey(db *database.BunDB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := new(model.Key)
		if err := c.BodyParser(key); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request",
			})
		}
		existingKey, err := db.GetKey(c.Context(), key.Kid)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.KeyErrorResponse(err))
		}
		if existingKey != nil {
			return c.Status(fiber.StatusConflict).JSON(presenter.KeyErrorResponse(errors.New("key already exists")))
		}
		if err := db.SaveKey(c.Context(), key); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(presenter.KeyErrorResponse(err))
		}
		return c.Status(200).JSON(presenter.KeySuccessResponse(key))
	}
}
