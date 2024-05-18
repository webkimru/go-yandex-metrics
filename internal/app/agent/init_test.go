package agent

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestSetup(t *testing.T) {
	t.Run("config with envs", func(t *testing.T) {
		os.Setenv("REPORT_INTERVAL", "10")
		os.Setenv("POLL_INTERVAL", "2")
		os.Setenv("KEY", "123")
		os.Setenv("RATE_LIMIT", "1")
		os.Setenv("ADDRESS", "localhost:8080")
		os.Setenv("CRYPTO_KEY", "123")
		os.Setenv("SERVER_PROTOCOL", "HTTP")

		_, _, err := Setup()
		assert.NoError(t, err)
	})
}
