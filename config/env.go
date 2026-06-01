package config

import (
	"os"
	"strings"
)

type envSource struct {
	dotenv []string
}

func (s *envSource) Load(target any) error {
	// Load .env files first (sets os.Environ).
	for _, path := range s.dotenv {
		if err := loadDotenvFile(path); err != nil {
			return err
		}
	}

	fields, err := collectFields(target)
	if err != nil {
		return err
	}

	for _, f := range fields {
		envName := envNameFromPath(f.path)
		val := os.Getenv(envName)
		if val == "" {
			continue
		}
		if err := setField(f.value, val); err != nil {
			return err
		}
	}
	return nil
}

func loadDotenvFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return parseDotenv(string(data))
}

func parseDotenv(content string) error {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if rest, ok := strings.CutPrefix(line, "export "); ok {
			line = strings.TrimSpace(rest)
		}

		eq := strings.IndexByte(line, '=')
		if eq < 0 {
			continue
		}

		key := strings.TrimSpace(line[:eq])
		val := strings.TrimSpace(line[eq+1:])
		val = unquote(val)

		if err := os.Setenv(key, val); err != nil {
			return err
		}
	}
	return nil
}

func unquote(s string) string {
	if len(s) < 2 {
		return s
	}
	switch s[0] {
	case '\'', '"':
		if s[len(s)-1] == s[0] {
			return s[1 : len(s)-1]
		}
	}
	return s
}
