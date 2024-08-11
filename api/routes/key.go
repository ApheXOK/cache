package routes

import (
	"cache/api/handler"
	"cache/database"
	"github.com/gofiber/fiber/v2"
)

func KeyRouter(app fiber.Router, db *database.BunDB) {
	k := app.Group("/api/keys")
	k.Post("/find", handler.GetKeys(db))
	k.Get("/:kid", handler.GetKey(db))
	k.Post("/", handler.SaveKey(db))
}
