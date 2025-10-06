package repository

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"time"
	"todo-service/pkg/domain"

	"errors"

	"github.com/jackc/pgx/v5"
)

type ISubscriptionsRepository interface {
	CreateSubscription(domain.Subscription) (int, error)
	GetSubscriptions() ([]domain.Subscription, error)
	GetSubscriptionByID(int) (*domain.Subscription, error)
	UpdateSubscription(int, domain.Subscription) error
	DeleteSubscription(int) error
	GetSubscriptionsForPeriod(periodStart, periodEnd time.Time, userID *uuid.UUID, serviceName *string) ([]domain.Subscription, error)
}

type SubscriptionsRepository struct {
	conn *pgx.Conn
}

func NewSubscriptionsRepository(conn *pgx.Conn) *SubscriptionsRepository {
	return &SubscriptionsRepository{conn: conn}
}

func (r *SubscriptionsRepository) CreateSubscription(s domain.Subscription) (int, error) {
	const op = "storage.repository.subscriptions.CreateSubscription"

	query := `
INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, now(), now())
RETURNING id
`
	var id int
	err := r.conn.QueryRow(context.Background(), query,
		s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate).Scan(&id)
	if err != nil {
		return -1, errors.New(fmt.Sprintf("error creating subscription: %w: %v", err, op))
	}
	return id, nil
}

func (r *SubscriptionsRepository) GetSubscriptions() ([]domain.Subscription, error) {
	const op = "storage.repository.subscriptions.GetSubscriptions"

	rows, err := r.conn.Query(context.Background(), `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at FROM subscriptions`)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error getting subscriptions: %w: %v", err, op))
	}
	defer rows.Close()

	var subs []domain.Subscription
	for rows.Next() {
		var s domain.Subscription
		if err := rows.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, errors.New(fmt.Sprintf("error scanning subscription: %w: %v", err, op))
		}
		subs = append(subs, s)
	}
	return subs, nil
}

func (r *SubscriptionsRepository) GetSubscriptionByID(id int) (*domain.Subscription, error) {
	const op = "storage.repository.subscriptions.GetSubscriptionByID"
	row := r.conn.QueryRow(context.Background(), `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at FROM subscriptions WHERE id = $1`, id)
	var s domain.Subscription
	if err := row.Scan(&s.ID, &s.ServiceName, &s.Price, &s.UserID, &s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return nil, errors.New(fmt.Sprintf("error getting subscription by id: %w: %v", err, op))
	}
	return &s, nil
}

func (r *SubscriptionsRepository) UpdateSubscription(id int, s domain.Subscription) error {
	const op = "storage.repository.subscriptions.UpdateSubscription"

	_, err := r.conn.Exec(context.Background(),
		`UPDATE subscriptions SET service_name=$1, price=$2, start_date=$3, end_date=$4, updated_at=now() WHERE id=$5`,
		s.ServiceName, s.Price, s.StartDate, s.EndDate, id)
	if err != nil {
		return errors.New(fmt.Sprintf("error updating subscription: %w: %v", err, op))
	}
	return nil
}

func (r *SubscriptionsRepository) DeleteSubscription(id int) error {
	const op = "storage.repository.subscriptions.DeleteSubscription"
	_, err := r.conn.Exec(context.Background(), `DELETE FROM subscriptions WHERE id = $1`, id)
	if err != nil {
		return errors.New(fmt.Sprintf("error deleting subscription: %w: %v", err, op))
	}
	return nil
}

// возвращает подписки, которые пересекаются с периодом
// фильтры userID и serviceName - опциональные
func (r *SubscriptionsRepository) GetSubscriptionsForPeriod(
	periodStart, periodEnd time.Time,
	userID *uuid.UUID, serviceName *string,
) ([]domain.Subscription, error) {

	const op = "storage.repository.subscriptions.GetSubscriptionsForPeriod"

	query := `
        SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
        FROM subscriptions
        WHERE start_date <= $1
          AND (end_date IS NULL OR end_date >= $2)
    `
	args := []interface{}{periodEnd, periodStart}

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", len(args)+1)
		args = append(args, *userID)
	}
	if serviceName != nil {
		query += fmt.Sprintf(" AND service_name = $%d", len(args)+1)
		args = append(args, *serviceName)
	}

	rows, err := r.conn.Query(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: query error: %w", op, err)
	}
	defer rows.Close()

	var subs []domain.Subscription
	for rows.Next() {
		var s domain.Subscription
		if err := rows.Scan(
			&s.ID, &s.ServiceName, &s.Price, &s.UserID,
			&s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("%s: scan error: %w", op, err)
		}
		subs = append(subs, s)
	}
	return subs, nil
}
