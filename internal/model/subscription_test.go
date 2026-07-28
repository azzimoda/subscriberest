package model_test

import (
	"testing"
	"time"

	"github.com/azzimoda/subscriberest/internal/model"
)

func TestParseSubscriptionDate(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		s       string
		want    time.Time
		wantErr bool
	}{
		{"2026-07 - Ok", "2026-07", time.Date(2026, 7, 0, 0, 0, 0, 0, time.UTC), false},
		{"0000-00 - Parsing error", "0000-00", time.Time{}, true},
		{"Invalid format", "invalida format", time.Time{}, true},
		{"Empty string", "", time.Time{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := model.ParseSubscriptionDate(tt.s)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ParseSubscriptionDate() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ParseSubscriptionDate() succeeded unexpectedly")
			}
			if got.Equal(tt.want) {
				t.Errorf("ParseSubscriptionDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubscription_IsActiveAt(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		s    model.Subscription
		// Named input parameters for target function.
		d    time.Time
		want bool
	}{
		{
			"2026-07 in 2026-07..2026-07 true",
			model.Subscription{
				StartDate: model.MustParseSubscriptionDate("2026-07"),
				EndDate:   new(model.MustParseSubscriptionDate("2026-07")),
			},
			model.MustParseSubscriptionDate("2026-07"),
			true,
		},
		{
			"2026-07 in 2026-07..2026-08 true",
			model.Subscription{
				StartDate: model.MustParseSubscriptionDate("2026-07"),
				EndDate:   new(model.MustParseSubscriptionDate("2026-08")),
			},
			model.MustParseSubscriptionDate("2026-07"),
			true,
		},
		{
			"2026-08 in 2026-07..2026-08 true",
			model.Subscription{
				StartDate: model.MustParseSubscriptionDate("2026-07"),
				EndDate:   new(model.MustParseSubscriptionDate("2026-08")),
			},
			model.MustParseSubscriptionDate("2026-08"),
			true,
		},
		{
			"2026-08 in 2026-07..2026-09 true",
			model.Subscription{
				StartDate: model.MustParseSubscriptionDate("2026-07"),
				EndDate:   new(model.MustParseSubscriptionDate("2026-09")),
			},
			model.MustParseSubscriptionDate("2026-08"),
			true,
		},
		{
			"2026-01 in 2026-07..2026-09 false",
			model.Subscription{
				StartDate: model.MustParseSubscriptionDate("2026-07"),
				EndDate:   new(model.MustParseSubscriptionDate("2026-09")),
			},
			model.MustParseSubscriptionDate("2026-01"),
			false,
		},
		{
			"2026-12 in 2026-07..2026-09 false",
			model.Subscription{
				StartDate: model.MustParseSubscriptionDate("2026-07"),
				EndDate:   new(model.MustParseSubscriptionDate("2026-09")),
			},
			model.MustParseSubscriptionDate("2026-12"),
			false,
		},
		{
			"2026-07 in 2026-07..nil true",
			model.Subscription{
				StartDate: model.MustParseSubscriptionDate("2026-07"),
				EndDate:   nil,
			},
			model.MustParseSubscriptionDate("2026-07"),
			true,
		},
		{
			"2026-08 in 2026-07..nil true",
			model.Subscription{
				StartDate: model.MustParseSubscriptionDate("2026-07"),
				EndDate:   nil,
			},
			model.MustParseSubscriptionDate("2026-08"),
			true,
		},
		{
			"2026-01 in 2026-07..nil true",
			model.Subscription{
				StartDate: model.MustParseSubscriptionDate("2026-07"),
				EndDate:   nil,
			},
			model.MustParseSubscriptionDate("2026-01"),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.IsActiveAt(tt.d)
			if got != tt.want {
				t.Errorf("IsActiveAt() = %v, want %v", got, tt.want)
			}
		})
	}
}
