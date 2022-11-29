package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jobstoit/tags/defaults"
	"github.com/jobstoit/tags/env"
)

type Config struct {
	Port        int    `env:"HTTPQ_PORT" default:"23411"`
	ServeTLS    bool   `env:"HTTPQ_SERVE_TLS" default:"true"`
	TLSKeyPath  string `env:"HTTPQ_TLS_KEY_PATH"`
	TLSCertPath string `env:"HTTPQ_TLS_CERT_PATH"`
}

func NewConfig() *Config {
	c := &Config{}

	if err := env.Parse(c); err != nil {
		log.Fatalf("unable to parse environent variables: %v", err)
	}

	if err := defaults.Parse(c); err != nil {
		log.Fatalf("unable to parse defaults in config: %v", err)
	}

	return c
}

func main() {
	config := NewConfig()
	httpq := &HTTPQ{}

	addr := fmt.Sprintf(":%d", config.Port)

	if config.ServeTLS {
		http.ListenAndServeTLS(addr, config.TLSKeyPath, config.TLSCertPath, httpq)
	} else {
		http.ListenAndServe(addr, httpq)
	}
}
