package http

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var RequestCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total_count",
		Help: "Total number of HTTP requests count",
	},
	[]string{"method", "path", "status", "code"},
)

var reg = prometheus.NewRegistry()

func init() {
	reg.MustRegister(
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(),
		RequestCounter)
}

func RequestCounterMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		RequestCounter.WithLabelValues(
			c.Request().Method,
			c.Request().URL.Path,
			http.StatusText(c.Response().Status),
			strconv.Itoa(c.Response().Status),
		).Inc()
		return next(c)
	}
}
