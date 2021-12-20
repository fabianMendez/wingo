package date_test

import (
	"testing"
	"time"

	"github.com/fabianMendez/wingo/pkg/date"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEffectiveDate(t *testing.T) {
	actual, err := date.Parse("2022-03-07T00:00:00.000+0000")
	require.NoError(t, err)
	assert.Equal(t, time.Date(2022, time.March, 7, 0, 0, 0, 0, time.UTC), *actual)
}
