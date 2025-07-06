package web_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	resp, err := http.Get(testServer.URL + apiEndpoints.Health)
	assert.NoError(t, err, "Failed to send GET request")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK")

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err, "Failed to read response body")

	expected := `{"status":"healthy"}`
	assert.Equal(t, expected, string(body), "Expected healthy status response")
}
