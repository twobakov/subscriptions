package routes

import (
	"todo-service/internal/handlers"
	"todo-service/internal/services"
	"todo-service/internal/storage/repository"
	"todo-service/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"

	// Swagger
	fiberSwagger "github.com/swaggo/fiber-swagger"
	_ "todo-service/docs"
)

func InitRoutes(conn *pgx.Conn) *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		return c.Next()
	})
	app.Use(logger.LoggerMiddleware())

	subRepo := repository.NewSubscriptionsRepository(conn)
	subService := services.NewSubscriptionsService(subRepo)
	subHandler := handlers.NewSubscriptionsHandler(subService)

	// Swagger route
	app.Get("/swagger/*", fiberSwagger.WrapHandler) // доступ к Swagger UI по /swagger/index.html

	api := app.Group("/api")
	{
		// subscriptions endpoints
		api.Get("/subscriptions/sum", subHandler.SumCost)
		api.Get("/subscriptions", subHandler.GetSubscriptions)
		api.Post("/subscriptions", subHandler.CreateSubscription)
		api.Put("/subscriptions/:id", subHandler.Update)
		api.Delete("/subscriptions/:id", subHandler.Delete)
		api.Get("/subscriptions/:id", subHandler.GetByID)
	}

	return app
}
