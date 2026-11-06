// Package config reads runtime settings from the environment.
package config

import "os"

// Config holds runtime settings.
type Config struct {
	PaymentsURL string
	Addr        string
}

// Load reads settings from the environment, with defaults for local runs.
func Load() Config {
	c := Config{
		PaymentsURL: os.Getenv("PAYMENTS_URL"),
		Addr:        os.Getenv("ADDR"),
	}
	if c.Addr == "" {
		c.Addr = ":8080"
	}
	return c
}
