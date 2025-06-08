package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	// PASSWORD_FILE is the environment variable that specifies the path to the password file.
	PASSWORD_FILE = "PASSWORD_FILE"
)

func ReadEnv() error {
	var envFilePath string

	if passwordFilePath, ok := os.LookupEnv(PASSWORD_FILE); ok {
		envFilePath = passwordFilePath
	} else {
		envFilePath = ".env"
	}

	file, err := os.Open(envFilePath)
	if err != nil {
		return fmt.Errorf("failed to open env file %s: %w", envFilePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || line[0] == '#' {
			continue // skip empty lines and comments
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			os.Setenv(parts[0], parts[1])
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading env file: %w", err)
	}

	return nil
}
