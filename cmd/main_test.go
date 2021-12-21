package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanFlightNumber(t *testing.T) {
	actual := cleanFlightNumber("P5-7013")
	assert.Equal(t, actual, "7013")
}
