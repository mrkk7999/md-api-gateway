package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Services map[string]ServiceConfig `yaml:"services"`
}

type ServiceConfig struct {
	Target string              `yaml:"target"`
	Routes map[string][]string `yaml:"routes"`
}

var AuthSettings AuthConfig

// LoadAuthConfig loads RBAC rules from auth.yml
func LoadAuthConfig() {
	data, err := os.ReadFile("../config/auth.yml")
	if err != nil {
		log.Fatalf("Failed to read auth.yml: %v", err)
	}

	err = yaml.Unmarshal(data, &AuthSettings)

	if err != nil {
		log.Fatalf("Failed to parse auth.yml: %v", err)
	}
}
