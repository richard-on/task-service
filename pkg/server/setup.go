package server

import (
	"github.com/richard-on/task-service/config"
	"github.com/richard-on/task-service/pkg/logger"
	"time"

	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
)

type Server struct {
	app *fiber.App
	log logger.Logger
}

func NewApp() Server {
	log := logger.NewLogger(config.DefaultWriter,
		config.LogInfo.Level,
		"mail-server")

	app := fiber.New(fiber.Config{
		Prefork:       config.FiberPrefork,
		ServerHeader:  "api.richardhere.dev",
		CaseSensitive: false,
		ReadTimeout:   time.Second * 30,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {

			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			err = ctx.SendStatus(code)
			if err != nil {
				// In case the SendFile fails
				return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
			}

			return nil
		},
	})

	prometheus := fiberprometheus.New("api.richardhere.dev/model")
	prometheus.RegisterAt(app, "/metrics")

	app.Use(
		//csrf.New(),
		cors.New(cors.ConfigDefault),
		recover.New(),
		pprof.New(
			pprof.Config{Next: func(c *fiber.Ctx) bool {
				return config.Env != "dev"
			}}),
		prometheus.Middleware,
		logger.Middleware(
			logger.NewLogger(config.DefaultWriter,
				config.LogInfo.Level,
				"model-httpserver"), nil,
		),
	)

	// Registering Swagger API
	app.Get("/swagger/*", swagger.HandlerDefault)

	return Server{
		app: app,
		log: log,
	}
}
