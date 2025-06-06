package config

import (
	"bufio"
	"os"
	"stream/pkg/logger"
	"strings"
)

func ReadEnv() {
	if _, err := os.Stat(".env"); err == nil {
		file, err := os.Open(".env")
		if err != nil {
			return
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
				logger.Debug.Printf("Setting environment variable: %s=%s", parts[0], parts[1])
				os.Setenv(parts[0], parts[1])
			}
		}
	}
}
