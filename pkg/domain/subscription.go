package domain

import (
	"github.com/google/uuid"
	"time"
)

type Subscription struct {
	ID          int
	ServiceName string
	Price       int        // целые рубли
	UserID      uuid.UUID  // UUID в виде строки
	StartDate   time.Time  // первый день месяца
	EndDate     *time.Time // nil = бессрочно
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
