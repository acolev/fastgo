package handlers

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var metricsHandler = adaptor.HTTPHandler(promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}))

func New() fiber.Handler {
	return metricsHandler
}

var _ http.Handler = promhttp.Handler()
