package routes

import (
	"github.com/gofiber/fiber/v2"

	"simple-job-tracker-backend/handlers"
	"simple-job-tracker-backend/middleware"
)

func Setup(app *fiber.App, jwtSecret string) {
	auth := &handlers.AuthHandler{JWTSecret: jwtSecret}
	job := &handlers.JobApplicationHandler{}

	app.Post("/api/auth/register", auth.Register)
	app.Post("/api/auth/login", auth.Login)

	api := app.Group("/api", middleware.AuthRequired(jwtSecret))
	api.Get("/auth/me", auth.Me)

	api.Get("/applications", job.List)
	api.Get("/applications/:id", job.Get)
	api.Post("/applications", job.Create)
	api.Put("/applications/:id", job.Update)
	api.Delete("/applications/:id", job.Delete)
}
