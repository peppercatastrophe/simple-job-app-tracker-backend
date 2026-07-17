package handlers

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"

	"simple-job-tracker-backend/database"
	"simple-job-tracker-backend/models"
)

type JobApplicationHandler struct{}

func (h *JobApplicationHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	rows, err := database.Pool.Query(
		context.Background(),
		`SELECT id, user_id, title, company, category, description, tech_stack, status, apply_date_time, created_at, updated_at
		 FROM job_applications WHERE user_id = $1 ORDER BY apply_date_time DESC NULLS LAST, created_at DESC`,
		userID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch applications"})
	}
	defer rows.Close()

	apps := []models.JobApplication{}
	for rows.Next() {
		var a models.JobApplication
		if err := rows.Scan(&a.ID, &a.UserID, &a.Title, &a.Company, &a.Category, &a.Description, &a.TechStack, &a.Status, &a.ApplyDateTime, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to scan application"})
		}
		apps = append(apps, a)
	}

	return c.JSON(apps)
}

func (h *JobApplicationHandler) Get(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	appID := c.Params("id")

	var a models.JobApplication
	err := database.Pool.QueryRow(
		context.Background(),
		`SELECT id, user_id, title, company, category, description, tech_stack, status, apply_date_time, created_at, updated_at
		 FROM job_applications WHERE id = $1 AND user_id = $2`,
		appID, userID,
	).Scan(&a.ID, &a.UserID, &a.Title, &a.Company, &a.Category, &a.Description, &a.TechStack, &a.Status, &a.ApplyDateTime, &a.CreatedAt, &a.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "application not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch application"})
	}

	return c.JSON(a)
}

func (h *JobApplicationHandler) Create(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var input models.CreateJobApplicationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if input.Title == "" || input.Company == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "title and company are required"})
	}

	status := input.Status
	if status == "" {
		status = "applied"
	}

	var applyDateTime *time.Time
	if input.ApplyDateTime != nil {
		t, err := time.Parse(time.RFC3339, *input.ApplyDateTime)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid apply_date_time format (use RFC3339)"})
		}
		applyDateTime = &t
	}

	var a models.JobApplication
	err := database.Pool.QueryRow(
		context.Background(),
		`INSERT INTO job_applications (user_id, title, company, category, description, tech_stack, status, apply_date_time)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, user_id, title, company, category, description, tech_stack, status, apply_date_time, created_at, updated_at`,
		userID, input.Title, input.Company, input.Category, input.Description, input.TechStack, status, applyDateTime,
	).Scan(&a.ID, &a.UserID, &a.Title, &a.Company, &a.Category, &a.Description, &a.TechStack, &a.Status, &a.ApplyDateTime, &a.CreatedAt, &a.UpdatedAt)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create application"})
	}

	eventDate := a.CreatedAt
	if applyDateTime != nil {
		eventDate = *applyDateTime
	}
	_, err = database.Pool.Exec(
		context.Background(),
		`INSERT INTO job_application_events (job_application_id, status, event_date) VALUES ($1, $2, $3)`,
		a.ID, a.Status, eventDate,
	)
	if err != nil {
		log.Printf("WARN failed to create initial event for application %s: %v", a.ID, err)
	}

	return c.Status(fiber.StatusCreated).JSON(a)
}

func (h *JobApplicationHandler) Update(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	appID := c.Params("id")

	var input models.UpdateJobApplicationInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	var previousStatus string
	_ = database.Pool.QueryRow(
		context.Background(),
		`SELECT status FROM job_applications WHERE id = $1 AND user_id = $2`,
		appID, userID,
	).Scan(&previousStatus)

	var applyDateTime *time.Time
	if input.ApplyDateTime != nil {
		t, err := time.Parse(time.RFC3339, *input.ApplyDateTime)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid apply_date_time format (use RFC3339)"})
		}
		applyDateTime = &t
	}

	var a models.JobApplication
	err := database.Pool.QueryRow(
		context.Background(),
		`UPDATE job_applications SET
			title = COALESCE($3, title),
			company = COALESCE($4, company),
			category = COALESCE($5, category),
			description = COALESCE($6, description),
			tech_stack = CASE WHEN $7::TEXT[] IS NOT NULL THEN $7 ELSE tech_stack END,
			status = COALESCE($8, status),
			apply_date_time = COALESCE($9, apply_date_time),
			updated_at = CURRENT_TIMESTAMP
		 WHERE id = $1 AND user_id = $2
		 RETURNING id, user_id, title, company, category, description, tech_stack, status, apply_date_time, created_at, updated_at`,
		appID, userID, input.Title, input.Company, input.Category, input.Description, input.TechStack, input.Status, applyDateTime,
	).Scan(&a.ID, &a.UserID, &a.Title, &a.Company, &a.Category, &a.Description, &a.TechStack, &a.Status, &a.ApplyDateTime, &a.CreatedAt, &a.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "application not found"})
		}
		log.Printf("ERROR updating application %s for user %s: %v", appID, userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update application"})
	}

	if input.Status != nil && *input.Status != previousStatus {
		eventDate := a.UpdatedAt
		if applyDateTime != nil {
			eventDate = *applyDateTime
		}
		_, evErr := database.Pool.Exec(
			context.Background(),
			`INSERT INTO job_application_events (job_application_id, status, event_date) VALUES ($1, $2, $3)`,
			a.ID, a.Status, eventDate,
		)
		if evErr != nil {
			log.Printf("WARN failed to create event for application %s: %v", a.ID, evErr)
		}
	}

	return c.JSON(a)
}

func (h *JobApplicationHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	appID := c.Params("id")

	tag, err := database.Pool.Exec(
		context.Background(),
		`DELETE FROM job_applications WHERE id = $1 AND user_id = $2`,
		appID, userID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete application"})
	}

	if tag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "application not found"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

type JobApplicationEventHandler struct{}

func (h *JobApplicationEventHandler) ownsApplication(ctx context.Context, appID, userID string) (bool, error) {
	var exists bool
	err := database.Pool.QueryRow(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM job_applications WHERE id = $1 AND user_id = $2)`,
		appID, userID,
	).Scan(&exists)
	return exists, err
}

func (h *JobApplicationEventHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	appID := c.Params("id")

	owns, err := h.ownsApplication(context.Background(), appID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to verify application"})
	}
	if !owns {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "application not found"})
	}

	rows, err := database.Pool.Query(
		context.Background(),
		`SELECT id, job_application_id, status, note, event_date, created_at
		 FROM job_application_events WHERE job_application_id = $1 ORDER BY event_date ASC, created_at ASC`,
		appID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch events"})
	}
	defer rows.Close()

	events := []models.JobApplicationEvent{}
	for rows.Next() {
		var e models.JobApplicationEvent
		if err := rows.Scan(&e.ID, &e.JobApplicationID, &e.Status, &e.Note, &e.EventDate, &e.CreatedAt); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to scan event"})
		}
		events = append(events, e)
	}

	return c.JSON(events)
}

func (h *JobApplicationEventHandler) Create(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	appID := c.Params("id")

	owns, err := h.ownsApplication(context.Background(), appID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to verify application"})
	}
	if !owns {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "application not found"})
	}

	var input models.CreateJobApplicationEventInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if input.Status == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status is required"})
	}

	eventDate := time.Now()
	if input.EventDate != nil {
		t, err := time.Parse(time.RFC3339, *input.EventDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid event_date format (use RFC3339)"})
		}
		eventDate = t
	}

	var e models.JobApplicationEvent
	err = database.Pool.QueryRow(
		context.Background(),
		`INSERT INTO job_application_events (job_application_id, status, note, event_date)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, job_application_id, status, note, event_date, created_at`,
		appID, input.Status, input.Note, eventDate,
	).Scan(&e.ID, &e.JobApplicationID, &e.Status, &e.Note, &e.EventDate, &e.CreatedAt)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create event"})
	}

	_, err = database.Pool.Exec(
		context.Background(),
		`UPDATE job_applications SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`,
		input.Status, appID,
	)
	if err != nil {
		log.Printf("WARN failed to sync application status for %s: %v", appID, err)
	}

	return c.Status(fiber.StatusCreated).JSON(e)
}

func (h *JobApplicationEventHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	appID := c.Params("id")
	eventID := c.Params("eventId")

	owns, err := h.ownsApplication(context.Background(), appID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to verify application"})
	}
	if !owns {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "application not found"})
	}

	tag, err := database.Pool.Exec(
		context.Background(),
		`DELETE FROM job_application_events WHERE id = $1 AND job_application_id = $2`,
		eventID, appID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete event"})
	}
	if tag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "event not found"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
