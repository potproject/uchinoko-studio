package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("go", "run", "github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0", "generate", "-f", "sqlc.yaml")
	cmd.Dir = "."
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
