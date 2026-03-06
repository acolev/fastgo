package bootstrap

import (
	"fastgo/internal/i18n"
	"fastgo/internal/shared/response"
	"time"

	"github.com/gofiber/fiber/v3"

	"fastgo/internal/config"
	"fastgo/internal/http/probes"
	httptests "fastgo/internal/http/tests"
)

func New(cfg *config.Config) *fiber.App {
	InitProviders(cfg)
	RunMigrations()
	app := fiber.New(fiber.Config{
		AppName:      "FastGo",
		ErrorHandler: response.ErrorHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	})

	i18n.SetDefaultLang("en")
	err := i18n.LoadDir("locales")
	if err != nil {
		panic(err)
	}
	app.Use(i18n.Middleware())

	api := app.Group("/api")

	testsGroup := api.Group("/t")
	httptests.RegisterRoutes(testsGroup)

	probesGroup := api.Group("/probes")
	probes.RegisterRoutes(probesGroup)

	return app
}
