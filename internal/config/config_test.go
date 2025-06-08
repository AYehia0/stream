package config

import (
	"os"
	"testing"
)

func writeEnvFile(t *testing.T, content string) {
	t.Helper()
	err := os.WriteFile(".env", []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write .env file: %v", err)
	}
}

func removeEnvFile(t *testing.T) {
	t.Helper()
	err := os.Remove(".env")
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to remove .env file: %v", err)
	}
}

func clearEnvVars(keys ...string) {
	for _, k := range keys {
		os.Unsetenv(k)
	}
}

func TestReadEnv_FileExists(t *testing.T) {
	envContent := `
# This is a comment
FOO=bar
EMPTY=
# Another comment
BAZ=qux
`
	writeEnvFile(t, envContent)
	defer removeEnvFile(t)
	clearEnvVars("FOO", "EMPTY", "BAZ")

	ReadEnv()

	if got := os.Getenv("FOO"); got != "bar" {
		t.Errorf("FOO = %q; want 'bar'", got)
	}
	if got := os.Getenv("EMPTY"); got != "" {
		t.Errorf("EMPTY = %q; want ''", got)
	}
	if got := os.Getenv("BAZ"); got != "qux" {
		t.Errorf("BAZ = %q; want 'qux'", got)
	}
}

func TestReadEnv_FileNotExists(t *testing.T) {
	removeEnvFile(t) // ensure no .env file
	clearEnvVars("FOO")

	ReadEnv()

	// Should not set anything or panic
	if got := os.Getenv("FOO"); got != "" {
		t.Errorf("FOO = %q; want '' when no .env file", got)
	}
}

func TestReadEnv_InvalidLines(t *testing.T) {
	envContent := `
INVALID_LINE_WITHOUT_EQUALS
# Comment line
ANOTHER=valid
`
	writeEnvFile(t, envContent)
	defer removeEnvFile(t)
	clearEnvVars("ANOTHER")

	ReadEnv()

	if got := os.Getenv("ANOTHER"); got != "valid" {
		t.Errorf("ANOTHER = %q; want 'valid'", got)
	}
}

func TestReadEnv_EmptyFile(t *testing.T) {
	writeEnvFile(t, "")
	defer removeEnvFile(t)

	ReadEnv()
}
