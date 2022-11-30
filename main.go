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
	httpq := NewHTTPQ()

	addr := fmt.Sprintf(":%d", config.Port)
	log.Printf("starting server on %s", addr)

	if config.TLSKeyPath != "" && config.TLSCertPath != "" {
		log.Fatal(http.ListenAndServeTLS(addr, config.TLSCertPath, config.TLSKeyPath, httpq))
	} else {
		log.Fatal(http.ListenAndServe(addr, httpq))
	}
}
