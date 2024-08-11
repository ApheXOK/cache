package presenter

import (
	"cache/model"
	"github.com/gofiber/fiber/v2"
)

func KeySuccessResponse(key *model.Key) *fiber.Map {
	return &fiber.Map{
		"status": true,
		"data":   key,
		"error":  nil,
	}
}

func KeyListSuccessResponse(keys []model.Key) *fiber.Map {
	return &fiber.Map{
		"status": true,
		"data":   keys,
		"error":  nil,
	}
}

func KeyErrorResponse(err error) *fiber.Map {
	return &fiber.Map{
		"status": false,
		"data":   "",
		"error":  err.Error(),
	}
}
