package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"simple-job-tracker-backend/config"
	"simple-job-tracker-backend/database"
	"simple-job-tracker-backend/routes"
)

func main() {
	cfg := config.Load()

	if err := database.Connect(cfg.DatabaseURL); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := database.RunMigrations(); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	app := fiber.New()

	app.Use(logger.New())
	app.Use(cors.New())

	routes.Setup(app, cfg.JWTSecret)

	log.Fatal(app.Listen(":" + cfg.Port))
}
