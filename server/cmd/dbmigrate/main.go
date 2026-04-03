package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/potproject/uchinoko-studio/db"
	"github.com/potproject/uchinoko-studio/envgen"
)

func main() {
	defaultDBPath := loadDefaultDBPath()

	dbPath := flag.String("db-path", defaultDBPath, "SQLite database path or directory")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [flags] <command> [args]\n", os.Args[0])
		fmt.Fprintln(flag.CommandLine.Output(), "")
		fmt.Fprintln(flag.CommandLine.Output(), "Examples:")
		fmt.Fprintln(flag.CommandLine.Output(), "  go run ./cmd/dbmigrate create add_user_setting")
		fmt.Fprintln(flag.CommandLine.Output(), "  go run ./cmd/dbmigrate up")
		fmt.Fprintln(flag.CommandLine.Output(), "  go run ./cmd/dbmigrate status")
		fmt.Fprintln(flag.CommandLine.Output(), "  go run ./cmd/dbmigrate down-to 1")
		fmt.Fprintln(flag.CommandLine.Output(), "")
		fmt.Fprintln(flag.CommandLine.Output(), "Commands are provided by goose. New files are created with zero-padded sequential numbers for sqlc compatibility.")
		fmt.Fprintln(flag.CommandLine.Output(), "")
		flag.PrintDefaults()
	}
	flag.Parse()

	if err := db.RunMigrationCommand(context.Background(), *dbPath, flag.Args()); err != nil {
		log.Fatal(err)
	}
}

func loadDefaultDBPath() string {
	if err := loadEnv(); err == nil {
		if path := strings.TrimSpace(envgen.Get().DB_FILE_PATH()); path != "" {
			return path
		}
	}
	return "database"
}

func loadEnv() error {
	envFile := ".env"
	if _, err := os.Stat(envFile); err != nil {
		envFile = "env.txt"
	}
	if err := godotenv.Load(envFile); err != nil {
		log.Println("Not loading .env file")
	}
	return envgen.Load()
}
