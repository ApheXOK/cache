package main

import (
	"cache/api/routes"
	"cache/database"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"os"
	"os/signal"
	"syscall"

	"time"
)

var (
	apiKey = "testabc123"
)

func validateAPIKey(_ *fiber.Ctx, key string) (bool, error) {
	hashedAPIKey := sha256.Sum256([]byte(apiKey))
	hashedKey := sha256.Sum256([]byte(key))

	if subtle.ConstantTimeCompare(hashedAPIKey[:], hashedKey[:]) == 1 {
		return true, nil
	}
	return false, keyauth.ErrMissingOrMalformedAPIKey
}

func main() {
	if envKey := os.Getenv("API_KEY"); envKey != "" {
		apiKey = envKey
	}
	app := fiber.New()
	app.Use(keyauth.New(keyauth.Config{
		SuccessHandler: func(c *fiber.Ctx) error {
			return c.Next()
		},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if errors.Is(err, keyauth.ErrMissingOrMalformedAPIKey) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "missing or malformed api key",
				})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid api key",
			})
		},
		KeyLookup: "cookie:api-key",
		Validator: validateAPIKey,
	}))

	app.Use(logger.New(logger.Config{
		CustomTags: map[string]logger.LogFunc{
			"ip": func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
				return output.WriteString(c.IP())
			},
		},
		Format:       "PID: ${pid} | IP: ${ip} | Time: [${time}] | Latency: ${latency} | Status: ${status} | ${method}| ${path}\n",
		TimeFormat:   "15:04:05 02-01-2006",
		TimeZone:     "Asia/Ho_Chi_Minh",
		TimeInterval: 500 * time.Millisecond,
	}))

	app.Use(recover.New())

	db, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}
	routes.KeyRouter(app, db)

	go func() {
		log.Fatal(app.Listen("0.0.0.0:5600"))
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	select {
	case <-quit:
		log.Info("Shutting down server...")
		_ = db.Close()
	}

}
