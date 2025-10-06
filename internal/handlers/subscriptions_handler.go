package handlers

import (
	"errors"
	"github.com/google/uuid"
	"log"
	"strings"
	"time"
	"todo-service/internal/dto"
	"todo-service/internal/services"
	"todo-service/pkg/domain"

	"github.com/gofiber/fiber/v2"
)

// SubscriptionsHandler handles subscriptions endpoints
type SubscriptionsHandler struct {
	service services.ISubscriptionsService
}

func NewSubscriptionsHandler(service services.ISubscriptionsService) *SubscriptionsHandler {
	return &SubscriptionsHandler{service: service}
}

// CreateSubscription
// @Summary Create subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body dto.CreateSubscriptionDTO true "Subscription"
// @Success 201 {object} dto.CreateSubscriptionResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/subscriptions [post]
func (h *SubscriptionsHandler) CreateSubscription(c *fiber.Ctx) error {
	const op = "handlers.subscriptions.CreateSubscription"

	var req dto.CreateSubscriptionDTO
	if err := c.BodyParser(&req); err != nil {
		log.Printf("%s: error parsing body: %v", op, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	start, err := parseMonthYear(req.StartDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid start_date format, expected MM-YYYY or YYYY-MM"})
	}
	var end *time.Time
	if req.EndDate != "" {
		t, err := parseMonthYear(req.EndDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid end_date format, expected MM-YYYY or YYYY-MM"})
		}
		end = &t
	}

	sub := domain.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   start,
		EndDate:     end,
	}

	id, err := h.service.CreateSubscription(sub)
	if err != nil {
		log.Printf("%s: error creating subscription: %v", op, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error creating subscription"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

// GetSubscriptions
// @Summary List subscriptions
// @Tags subscriptions
// @Produce json
// @Success 200 {array} domain.Subscription
// @Router /api/subscriptions [get]
func (h *SubscriptionsHandler) GetSubscriptions(c *fiber.Ctx) error {
	subs, err := h.service.GetSubscriptions()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"subscriptions": subs})
}

// GetByID
// @Summary Get subscription by id
// @Tags subscriptions
// @Produce json
// @Param id path int true "subscription id"
// @Success 200 {object} domain.Subscription
// @Failure 404 {object} map[string]string
// @Router /api/subscriptions/{id} [get]
func (h *SubscriptionsHandler) GetByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	sub, err := h.service.GetSubscriptionByID(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(sub)
}

// Update
// @Summary Update subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path int true "subscription id"
// @Param subscription body dto.UpdateSubscriptionDTO true "Subscription update"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]interface{}
// @Router /api/subscriptions/{id} [put]
func (h *SubscriptionsHandler) Update(c *fiber.Ctx) error {
	const op = "handlers.subscriptions.Update"

	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var req dto.UpdateSubscriptionDTO
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	// load existing to fill unchanged fields
	existing, err := h.service.GetSubscriptionByID(id)
	if err != nil || existing == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	}

	// patch fields
	if req.ServiceName != nil {
		existing.ServiceName = *req.ServiceName
	}
	if req.Price != nil {
		existing.Price = *req.Price
	}
	if req.StartDate != nil {
		t, err := parseMonthYear(*req.StartDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid start_date"})
		}
		existing.StartDate = t
	}
	if req.EndDate != nil {
		if *req.EndDate == "" {
			existing.EndDate = nil
		} else {
			t, err := parseMonthYear(*req.EndDate)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid end_date"})
			}
			existing.EndDate = &t
		}
	}

	if err := h.service.UpdateSubscription(id, *existing); err != nil {
		log.Printf("%s: error updating subscription: %v", op, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error updating subscription"})
	}

	return c.JSON(fiber.Map{"message": "updated"})
}

// Delete
// @Summary Delete subscription
// @Tags subscriptions
// @Param id path int true "subscription id"
// @Success 200 {object} map[string]string
// @Router /api/subscriptions/{id} [delete]
func (h *SubscriptionsHandler) Delete(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := h.service.DeleteSubscription(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error deleting"})
	}
	return c.JSON(fiber.Map{"message": "deleted"})
}

// SumCostResponse represents total sum response
type SumCostResponse struct {
	Total int `json:"total"`
}

// SumCost handler
// @Summary Sum subscriptions cost for period
// @Tags subscriptions
// @Produce json
// @Param from query string true "from month" example(2025-07 or 07-2025)
// @Param to query string true "to month" example(2025-09)
// @Param user_id query string false "user id (uuid)"
// @Param service_name query string false "service name"
// @Success 200 {object} SumCostResponse
// @Router /api/subscriptions/sum [get]
func (h *SubscriptionsHandler) SumCost(c *fiber.Ctx) error {
	from := c.Query("from")
	to := c.Query("to")
	if from == "" || to == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "from and to are required (format YYYY-MM or MM-YYYY)"})
	}
	fromDate, err := parseMonthYear(from)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid from date"})
	}
	toDate, err := parseMonthYear(to)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid to date"})
	}

	periodStart := time.Date(fromDate.Year(), fromDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(toDate.Year(), toDate.Month(), 1, 0, 0, 0, 0, time.UTC)

	/*var userID *string
	if v := c.Query("user_id"); v != "" {
		userID = &v
	}*/
	var userUUID *uuid.UUID
	if v := c.Query("user_id"); v != "" {
		v = strings.Trim(v, "\" ")
		log.Printf("user_id raw: '%s'", v)
		parsed, err := uuid.Parse(v)
		if err != nil {
			log.Printf("uuid.Parse error: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user_id"})
		}
		userUUID = &parsed
	}
	var serviceName *string
	if v := c.Query("service_name"); v != "" {
		serviceName = &v
	}

	total, err := h.service.SumCost(periodStart, periodEnd, userUUID, serviceName)
	if err != nil {
		log.Printf("handlers.subscriptions.SumCost: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "error calculating sum"})
	}

	return c.JSON(SumCostResponse{Total: int(total)})
}

// parseMonthYear accepts "MM-YYYY" or "YYYY-MM" and returns time on first day of that month in UTC.
func parseMonthYear(s string) (time.Time, error) {
	layouts := []string{"01-2006", "2006-01"}
	var t time.Time
	var err error
	for _, l := range layouts {
		t, err = time.Parse(l, s)
		if err == nil {
			// normalize to first day
			return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
		}
	}
	return time.Time{}, errors.New("invalid format")
}
