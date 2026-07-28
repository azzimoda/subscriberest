package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Service   string     `gorm:"type:varchar(100);not null" json:"service"`
	Price     int        `gorm:"not null" json:"price"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	StartDate time.Time  `gorm:"type:date;not null" json:"start_date"`
	EndDate   *time.Time `gorm:"type:date" json:"end_date,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (s Subscription) IsActiveAt(d time.Time) bool {
	return d.Equal(s.StartDate) ||
		(s.EndDate != nil && d.Equal(*s.EndDate)) ||
		d.After(s.StartDate) && (s.EndDate == nil || d.Before(*s.EndDate))
}

func MustParseSubscriptionDate(s string) time.Time {
	if d, err := ParseSubscriptionDate(s); err != nil {
		panic(err)
	} else {
		return d
	}
}
func ParseSubscriptionDate(s string) (date time.Time, err error) {
	if t, err := time.Parse("2006-01", s); err != nil {
		return time.Time{}, errors.New("invalid layout, expected 2006-01")
	} else {
		return t, nil
	}
}
