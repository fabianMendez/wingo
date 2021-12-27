package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanFlightNumber(t *testing.T) {
	actual := cleanFlightNumber("P5-7013")
	assert.Equal(t, actual, "7013")
}

func TestWetag(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "routes",
			path:     "testdata/routes.json",
			expected: `W/"2370-NeiDQizESMGkVpWJE/FiPtto0E8"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ioutil.ReadFile(tt.path)
			assert.NoError(t, err)
			actual := wetag(data)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

// reference: https://github.com/jshttp/etag/blob/4664b6e53c85a56521076f9c5004dd9626ae10c8/index.js
func wetag(buf []byte) string {
	if len(buf) == 0 {
		// fast-path empty body
		return `W/"0-0"`
	}

	hashBuf := sha1.Sum(buf)
	hash := base64.StdEncoding.EncodeToString(hashBuf[:])

	return fmt.Sprintf(`W/"%x-%s"`, len(buf), hash[0:27])
}
