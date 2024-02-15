package configuration

import (
	"flag"
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

// Env is environment variables
var (
	Env = ""
)

func init() {
	flag.StringVar(&Env, "env", "dev", "Set environment")
	flag.Parse()

	var errEnv error
	if Env != "dev" {
		fmt.Printf("Load .env.%s ...\n", Env)
		errEnv = godotenv.Load(fmt.Sprintf(".env.%s", Env))
	} else {
		fmt.Printf("Load .env ...\n")
		errEnv = godotenv.Load(".env.dev")
	}
	if errEnv != nil {
		log.Fatal("Error loading .env.dev file")
	}
	fmt.Println("Load environment file completed")
}

func Load() {
	fmt.Println("load configuration completed")
}
