package notifications_test

import (
	"testing"
	"time"

	"github.com/fabianMendez/wingo/pkg/notifications"
	"github.com/stretchr/testify/assert"
)

func TestFilterBetweenDates(t *testing.T) {
	tests := []struct {
		name          string
		subscriptions []notifications.Setting
		expected      []notifications.Setting
		start, end    time.Time
	}{
		{
			name: "all in range",
			subscriptions: []notifications.Setting{
				{Date: "2021-12-24"},
				{Date: "2021-12-25"},
			},
			expected: []notifications.Setting{
				{Date: "2021-12-24"},
				{Date: "2021-12-25"},
			},
			start: time.Date(2021, 12, 24, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2021, 12, 26, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "all after range",
			subscriptions: []notifications.Setting{
				{Date: "2021-12-24"},
				{Date: "2021-12-25"},
			},
			expected: []notifications.Setting{},
			start:    time.Date(2021, 12, 20, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2021, 12, 23, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "all before range",
			subscriptions: []notifications.Setting{
				{Date: "2021-11-24"},
				{Date: "2021-11-25"},
			},
			expected: []notifications.Setting{},
			start:    time.Date(2021, 12, 20, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2021, 12, 23, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "some after range",
			subscriptions: []notifications.Setting{
				{Date: "2021-12-24"},
				{Date: "2021-12-25"},
			},
			expected: []notifications.Setting{
				{Date: "2021-12-24"},
			},
			start: time.Date(2021, 12, 20, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2021, 12, 25, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "some before range",
			subscriptions: []notifications.Setting{
				{Date: "2021-12-24"},
				{Date: "2021-12-25"},
			},
			expected: []notifications.Setting{
				{Date: "2021-12-25"},
			},
			start: time.Date(2021, 12, 25, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2021, 12, 26, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := notifications.FilterBetweenDates(tt.subscriptions, tt.start, tt.end)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
