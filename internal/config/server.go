package config

import (
	"flag"
	"fmt"
	"os"
	"regexp"
)

type ServerConfig struct {
	Address string
}

func NewServerConfig() *ServerConfig {
	cfg := &ServerConfig{
		Address: ":8080",
	}
	cfg.parseFlags()
	return cfg
}

func (sc *ServerConfig) parseFlags() {
	flag.StringVar(&sc.Address, "a", sc.Address, "Server address")
	flag.Parse()

	// Проверяем наличие схемы, если нет, то добавляем
	if idx := regexp.MustCompile(`.*?:\/\/`).FindStringIndex(sc.Address); idx == nil {
		if idx := regexp.MustCompile(`^:`).FindStringIndex(sc.Address); idx == nil {
			sc.Address = "http://" + sc.Address
		}
	}

	// Проверяем, что не переданы неизвестные флаги
	if flag.NArg() > 0 {
		fmt.Printf("Error: unknown flags or arguments: %v\n", flag.Args())
		flag.Usage()
		os.Exit(1)
	}

}
