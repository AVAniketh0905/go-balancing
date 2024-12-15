package server

import "fmt"

type Configuration interface {
	Port() int
	Addr() string
	ParseFlags()
}

type Config struct {
	port int
}

func NewConfig(defaultPort int) Config {
	return Config{
		port: defaultPort,
	}
}

func (cfg *Config) Port() int { return cfg.port }

func (cfg *Config) Addr() string { return fmt.Sprintf("0.0.0.0:%v", cfg.port) }

func (cfg *Config) ParseFlags() {
	// flag.IntVar(&cfg.port, "port", cfg.port, "port to listen on")
	// flag.Parse()
}
