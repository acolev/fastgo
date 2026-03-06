package bootstrap

import (
	"fastgo/internal/i18n"
	"fastgo/internal/shared/logger"
	"fastgo/internal/shared/response"
	"fmt"
	"time"

	fiberswagger "github.com/gofiber/contrib/v3/swagger"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"

	"fastgo/internal/config"
	"fastgo/internal/http/probes"
	httptests "fastgo/internal/http/tests"
)

func New(cfg *config.Config) (*fiber.App, error) {
	const swaggerFilePath = "docs/swagger.json"

	if err := InitProviders(cfg); err != nil {
		return nil, err
	}

	needsProviderCleanup := true
	defer func() {
		if needsProviderCleanup {
			_ = ShutdownProviders()
		}
	}()

	if err := RunMigrations(); err != nil {
		return nil, err
	}

	app := fiber.New(fiber.Config{
		AppName:      cfg.APP_NAME,
		ErrorHandler: response.ErrorHandler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	})

	i18n.SetDefaultLang("en")
	err := i18n.LoadDir("locales")
	if err != nil {
		return nil, fmt.Errorf("load locales: %w", err)
	}

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(i18n.Middleware())
	app.Use(logger.HTTPMiddleware(cfg.APP_ENV))

	swaggerSpec, err := loadSwaggerSpec(swaggerFilePath, cfg.APP_NAME)
	if err != nil {
		return nil, err
	}
	app.Get("/api/docs/swagger.json", swaggerSpecHandler(swaggerSpec))
	app.Use(fiberswagger.New(fiberswagger.Config{
		BasePath: "/",
		FilePath: swaggerFilePath,
		Path:     "docs",
		Title:    cfg.APP_NAME + " API Docs",
	}))

	api := app.Group("/api")
	v1 := api.Group("/v1")

	testsGroup := v1.Group("/t")
	httptests.RegisterRoutes(testsGroup)

	probes.RegisterRoutes(app)

	needsProviderCleanup = false

	return app, nil
}
