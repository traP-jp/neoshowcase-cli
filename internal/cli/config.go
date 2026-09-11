package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultAuthHeader = "X-Showcase-User"

type config struct {
	Endpoint   string `json:"endpoint"`
	User       string `json:"user"`
	AuthHeader string `json:"auth_header"`
}

func loadConfig(path string, explicitlySet bool) (config, error) {
	if path == "" {
		base, err := os.UserConfigDir()
		if err != nil {
			if explicitlySet {
				return config{}, err
			}
			return config{}, nil
		}
		path = filepath.Join(base, "neoshowcase-cli", "config.json")
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) && !explicitlySet {
		return config{}, nil
	}
	if err != nil {
		return config{}, fmt.Errorf("read config %q: %w", path, err)
	}
	var result config
	if json.Unmarshal(b, &result) == nil {
		return result, nil
	}
	// A deliberately small TOML-compatible parser keeps the configuration
	// human-friendly without adding a second configuration dependency.
	s := bufio.NewScanner(strings.NewReader(string(b)))
	for s.Scan() {
		line := strings.TrimSpace(strings.SplitN(s.Text(), "#", 2)[0])
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return config{}, fmt.Errorf("parse config %q: expected key = value", path)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
		switch key {
		case "endpoint":
			result.Endpoint = value
		case "user":
			result.User = value
		case "auth_header":
			result.AuthHeader = value
		default:
			return config{}, fmt.Errorf("parse config %q: unknown key %q", path, key)
		}
	}
	if err := s.Err(); err != nil {
		return config{}, fmt.Errorf("parse config %q: %w", path, err)
	}
	return result, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
