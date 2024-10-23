package admina

import (
	"os"
	"strconv"
	"testing"
)

func getTestDebugMode() bool {
	debugMode, _ := strconv.ParseBool(os.Getenv("ADMINA_TEST_DEBUG"))
	return debugMode
}

func TestNewClient(t *testing.T) {
	// Set environment variables
	os.Setenv("ADMINA_ORGANIZATION_ID", "test-org")
	os.Setenv("ADMINA_API_KEY", "test-key")
	defer func() {
		os.Unsetenv("ADMINA_ORGANIZATION_ID")
		os.Unsetenv("ADMINA_API_KEY")
	}()

	client := NewClient(getTestDebugMode())

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.BaseURL != DefaultBaseURL {
		t.Errorf("BaseURL mismatch. Expected: %s, Got: %s", DefaultBaseURL, client.BaseURL)
	}

	if client.Debug != getTestDebugMode() {
		t.Errorf("Debug mode mismatch. Expected: %v, Got: %v", getTestDebugMode(), client.Debug)
	}

	// Add more field checks as needed
}

// Add more test functions as needed
