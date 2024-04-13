package main

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/potproject/uchinoko-studio/db"
	"github.com/potproject/uchinoko-studio/envgen"
	"github.com/potproject/uchinoko-studio/router"
	"tailscale.com/tsnet"
)

func main() {
	envSetup()
	dbSetup()
	if envgen.Get().TAILSCALE_ENABLED() == false {
		serverSetup()
	} else {
		tailscaleSetup()
	}
}

func dbSetup() {
	db.Start()
}

func envSetup() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// Setup envgen package from environment variables
	err = envgen.Load()
	if err != nil {
		log.Fatal(err)
	}
}

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}

}

func serverSetup() {
	openBrowser("http://" + envgen.Get().HOST() + ":" + strconv.FormatInt(int64(envgen.Get().PORT()), 10))
	app := router.Route()
	log.Fatal(app.Listen(envgen.Get().HOST() + ":" + strconv.FormatInt(int64(envgen.Get().PORT()), 10)))
}

func tailscaleSetup() {
	s := new(tsnet.Server)
	s.Hostname = envgen.Get().TAILSCALE_HOSTNAME()
	addr := ":" + strconv.FormatInt(int64(envgen.Get().TAILSCALE_PORT()), 10)
	defer s.Close()

	var ln net.Listener
	var err error

	if envgen.Get().TAILSCALE_ENABLED_TLS() {
		if envgen.Get().TAILSCALE_FUNNEL_ENABLED() {
			ln, err = s.ListenFunnel("tcp", addr)
		} else {
			ln, err = s.ListenTLS("tcp", addr)
		}
	} else {
		ln, err = s.Listen("tcp", addr)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	app := router.Route()
	protocol := "http"
	if envgen.Get().TAILSCALE_ENABLED_TLS() {
		protocol = "https"
	}
	openBrowser(protocol + "://" + envgen.Get().TAILSCALE_HOSTNAME() + ":" + strconv.FormatInt(int64(envgen.Get().TAILSCALE_PORT()), 10))
	log.Fatal(app.Listener(ln))
}
