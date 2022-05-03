package main

import (
	"log"
	"os"
)

var envs = []string{"PORT", "AMQP_URL", "ENV"}

func checkEnvVars() {
	for _, env := range envs {
		if os.Getenv(env) == "" {
			log.Fatalln("environment variable", env, "is not specified but is required")
		}
	}
}
