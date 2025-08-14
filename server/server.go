package server

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"proxytrack/api"
	"proxytrack/middleware"
	"proxytrack/store"
	"strconv"
	"time"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var config = fiber.Config{
	ErrorHandler: api.ErrorHandler,
}

type Server struct {
	listenAddr string
	logger     *slog.Logger
}

func NewServer(addr string) *Server {
	return &Server{
		listenAddr: addr,
		logger:     slog.Default(),
	}
}

func (s *Server) Stop() {
	s.logger.Info("server stopped")
}

func RegisterMetrics(app *fiber.App) {
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
}

func (s *Server) Run() {
	url := os.Getenv("TARGET_URL")
	intervalStr := os.Getenv("REQ_TIMEOUT")
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		log.Fatal("Error to set ticker interval")
	}

	port, _ := strconv.Atoi(os.Getenv("PG_PORT"))
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", os.Getenv("PG_HOST"), port, os.Getenv("PG_USER"), os.Getenv("PG_PASS"), os.Getenv("PG_DB_NAME"))
	db, err := store.NewPostgresStore(connStr)
	if err != nil {
		s.logger.Error("error to connect to Posgres database", "error", err.Error())
		return
	}

	if err := db.Init(); err != nil {
		s.logger.Error("error to create tables", "error", err.Error())
		return
	}

	var (
		app          = fiber.New(config)
		ProxyHandler = api.NewProxyHandler(db, interval, url)
		promMetrics  = middleware.NewPromMetrics()
		apiv1        = app.Group("/TTPS/v1/service")
	)
	RegisterMetrics(app)

	apiv1.All("/*", WrapHandler(promMetrics, ProxyHandler.Proxy, "Proxy"))

	err = app.Listen(s.listenAddr)
	if err != nil {
		s.logger.Error("error to start server", "error", err.Error())
		return
	}
}

func WithLogging(handler fiber.Handler) fiber.Handler {
	return middleware.LoggingHandlerDecorator(handler)
}

func WrapHandler(p *middleware.PromMetrics, handler fiber.Handler, handlerName string) fiber.Handler {
	return p.WithMetrics(WithLogging(handler), handlerName)
}
