package services

import (
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"
	"todo-service/internal/storage/repository"
	"todo-service/pkg/domain"
)

type ISubscriptionsService interface {
	CreateSubscription(domain.Subscription) (int, error)
	GetSubscriptions() ([]domain.Subscription, error)
	GetSubscriptionByID(int) (*domain.Subscription, error)
	UpdateSubscription(int, domain.Subscription) error
	DeleteSubscription(int) error
	SumCost(periodStart, periodEnd time.Time, userID *uuid.UUID, serviceName *string) (int64, error)
}

type SubscriptionsService struct {
	repo repository.ISubscriptionsRepository
}

func NewSubscriptionsService(repo repository.ISubscriptionsRepository) *SubscriptionsService {
	return &SubscriptionsService{repo: repo}
}

func (s *SubscriptionsService) CreateSubscription(sub domain.Subscription) (int, error) {
	const op = "services.subscriptions_service.CreateSubscription"

	// Валидация
	if strings.TrimSpace(sub.ServiceName) == "" {
		return -1, fmt.Errorf("service_name cannot be empty: %v", op)
	}
	if sub.Price < 0 {
		return -1, fmt.Errorf("price must be >= 0: %v", op)
	}
	if sub.EndDate != nil {
		if sub.EndDate.Before(sub.StartDate) {
			return -1, fmt.Errorf("end_date cannot be before start_date: %v", op)
		}
	}

	return s.repo.CreateSubscription(sub)
}

func (s *SubscriptionsService) GetSubscriptions() ([]domain.Subscription, error) {
	return s.repo.GetSubscriptions()
}

func (s *SubscriptionsService) GetSubscriptionByID(id int) (*domain.Subscription, error) {
	return s.repo.GetSubscriptionByID(id)
}

func (s *SubscriptionsService) UpdateSubscription(id int, sub domain.Subscription) error {
	if sub.Price < 0 {
		return fmt.Errorf("price must be >= 0")
	}
	if sub.EndDate != nil && sub.EndDate.Before(sub.StartDate) {
		return fmt.Errorf("end_date cannot be before start_date")
	}
	return s.repo.UpdateSubscription(id, sub)
}

func (s *SubscriptionsService) DeleteSubscription(id int) error {
	return s.repo.DeleteSubscription(id)
}

// SumCost - суммарная стоимость за период (учитываем количество месяцев перекрытия)
// возвращает сумму в целых рублях (int64), чтобы избежать переполнения при больших суммах
func (s *SubscriptionsService) SumCost(
	periodStart, periodEnd time.Time,
	userID *uuid.UUID, serviceName *string,
) (int64, error) {
	const op = "services.subscriptions_service.SumCost"

	subs, err := s.repo.GetSubscriptionsForPeriod(periodStart, periodEnd, userID, serviceName)
	if err != nil {
		return 0, fmt.Errorf("error getting subs for period: %w: %v", err, op)
	}

	var total int64
	for _, sub := range subs {
		effStart := sub.StartDate
		effEnd := periodEnd
		if sub.EndDate != nil && sub.EndDate.Before(periodEnd) {
			effEnd = *sub.EndDate
		}
		if effStart.Before(periodStart) {
			effStart = periodStart
		}
		effStart = time.Date(effStart.Year(), effStart.Month(), 1, 0, 0, 0, 0, time.UTC)
		effEnd = time.Date(effEnd.Year(), effEnd.Month(), 1, 0, 0, 0, 0, time.UTC)

		months := monthsBetweenInclusive(effStart, effEnd) + 1
		if months < 0 {
			months = 0
		}
		total += int64(sub.Price) * int64(months)
	}

	return total, nil
}

// вспомогательная функция: количество полных месяцев между a и b, где a и b первый день месяца
func monthsBetweenInclusive(a, b time.Time) int {
	ay, am, _ := a.Date()
	by, bm, _ := b.Date()
	return (by-ay)*12 + int(bm-am)
}
