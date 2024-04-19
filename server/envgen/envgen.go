// Code generated by github.com/potproject/goenvgen, DO NOT EDIT.

package envgen

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

func (g getter) ANTHROPIC_API_KEY() string {
	return env.ANTHROPIC_API_KEY
}
func (s setter) ANTHROPIC_API_KEY(value string) {
	env.ANTHROPIC_API_KEY = value
	return
}
func (g getter) BERTVITS2_ENDPOINT() string {
	return env.BERTVITS2_ENDPOINT
}
func (s setter) BERTVITS2_ENDPOINT(value string) {
	env.BERTVITS2_ENDPOINT = value
	return
}
func (g getter) COHERE_API_KEY() string {
	return env.COHERE_API_KEY
}
func (s setter) COHERE_API_KEY(value string) {
	env.COHERE_API_KEY = value
	return
}
func (g getter) DB_FILE_PATH() string {
	return env.DB_FILE_PATH
}
func (s setter) DB_FILE_PATH(value string) {
	env.DB_FILE_PATH = value
	return
}
func (g getter) HOST() string {
	return env.HOST
}
func (s setter) HOST(value string) {
	env.HOST = value
	return
}
func (g getter) OPENAI_API_KEY() string {
	return env.OPENAI_API_KEY
}
func (s setter) OPENAI_API_KEY(value string) {
	env.OPENAI_API_KEY = value
	return
}
func (g getter) PORT() int32 {
	return env.PORT
}
func (s setter) PORT(value int32) {
	env.PORT = value
	return
}
func (g getter) STYLEBERTVIT2_ENDPOINT() string {
	return env.STYLEBERTVIT2_ENDPOINT
}
func (s setter) STYLEBERTVIT2_ENDPOINT(value string) {
	env.STYLEBERTVIT2_ENDPOINT = value
	return
}
func (g getter) TAILSCALE_ENABLED() bool {
	return env.TAILSCALE_ENABLED
}
func (s setter) TAILSCALE_ENABLED(value bool) {
	env.TAILSCALE_ENABLED = value
	return
}
func (g getter) TAILSCALE_ENABLED_TLS() bool {
	return env.TAILSCALE_ENABLED_TLS
}
func (s setter) TAILSCALE_ENABLED_TLS(value bool) {
	env.TAILSCALE_ENABLED_TLS = value
	return
}
func (g getter) TAILSCALE_FUNNEL_ENABLED() bool {
	return env.TAILSCALE_FUNNEL_ENABLED
}
func (s setter) TAILSCALE_FUNNEL_ENABLED(value bool) {
	env.TAILSCALE_FUNNEL_ENABLED = value
	return
}
func (g getter) TAILSCALE_HOSTNAME() string {
	return env.TAILSCALE_HOSTNAME
}
func (s setter) TAILSCALE_HOSTNAME(value string) {
	env.TAILSCALE_HOSTNAME = value
	return
}
func (g getter) TAILSCALE_PORT() int32 {
	return env.TAILSCALE_PORT
}
func (s setter) TAILSCALE_PORT(value int32) {
	env.TAILSCALE_PORT = value
	return
}
func (g getter) VOICEVOX_ENDPOINT() string {
	return env.VOICEVOX_ENDPOINT
}
func (s setter) VOICEVOX_ENDPOINT(value string) {
	env.VOICEVOX_ENDPOINT = value
	return
}

type environment struct {
	ANTHROPIC_API_KEY        string
	BERTVITS2_ENDPOINT       string
	COHERE_API_KEY           string
	DB_FILE_PATH             string
	HOST                     string
	OPENAI_API_KEY           string
	PORT                     int32
	STYLEBERTVIT2_ENDPOINT   string
	TAILSCALE_ENABLED        bool
	TAILSCALE_ENABLED_TLS    bool
	TAILSCALE_FUNNEL_ENABLED bool
	TAILSCALE_HOSTNAME       string
	TAILSCALE_PORT           int32
	VOICEVOX_ENDPOINT        string
}

var env environment

// Load reads the environment variables and stores them in the env variable.
// If the type conversion fails, it returns error.
func Load() error {
	var err error
	ANTHROPIC_API_KEY := os.Getenv("ANTHROPIC_API_KEY")
	BERTVITS2_ENDPOINT := os.Getenv("BERTVITS2_ENDPOINT")
	COHERE_API_KEY := os.Getenv("COHERE_API_KEY")
	DB_FILE_PATH := os.Getenv("DB_FILE_PATH")
	HOST := os.Getenv("HOST")
	OPENAI_API_KEY := os.Getenv("OPENAI_API_KEY")
	PORT__S := os.Getenv("PORT")
	PORT__64, err := strconv.ParseInt(PORT__S, 10, 32)
	if err != nil {
		return errors.New("PORT: " + err.Error())
	}
	PORT := int32(PORT__64)
	STYLEBERTVIT2_ENDPOINT := os.Getenv("STYLEBERTVIT2_ENDPOINT")
	TAILSCALE_ENABLED := false
	TAILSCALE_ENABLED__S := os.Getenv("TAILSCALE_ENABLED")
	if strings.ToLower(TAILSCALE_ENABLED__S) == "true" {
		TAILSCALE_ENABLED = true
	} else if strings.ToLower(TAILSCALE_ENABLED__S) == "false" {
		TAILSCALE_ENABLED = false
	} else {
		return errors.New("TAILSCALE_ENABLED: " + "cannot use " + TAILSCALE_ENABLED__S + " as type bool in assignment")
	}
	TAILSCALE_ENABLED_TLS := false
	TAILSCALE_ENABLED_TLS__S := os.Getenv("TAILSCALE_ENABLED_TLS")
	if strings.ToLower(TAILSCALE_ENABLED_TLS__S) == "true" {
		TAILSCALE_ENABLED_TLS = true
	} else if strings.ToLower(TAILSCALE_ENABLED_TLS__S) == "false" {
		TAILSCALE_ENABLED_TLS = false
	} else {
		return errors.New("TAILSCALE_ENABLED_TLS: " + "cannot use " + TAILSCALE_ENABLED_TLS__S + " as type bool in assignment")
	}
	TAILSCALE_FUNNEL_ENABLED := false
	TAILSCALE_FUNNEL_ENABLED__S := os.Getenv("TAILSCALE_FUNNEL_ENABLED")
	if strings.ToLower(TAILSCALE_FUNNEL_ENABLED__S) == "true" {
		TAILSCALE_FUNNEL_ENABLED = true
	} else if strings.ToLower(TAILSCALE_FUNNEL_ENABLED__S) == "false" {
		TAILSCALE_FUNNEL_ENABLED = false
	} else {
		return errors.New("TAILSCALE_FUNNEL_ENABLED: " + "cannot use " + TAILSCALE_FUNNEL_ENABLED__S + " as type bool in assignment")
	}
	TAILSCALE_HOSTNAME := os.Getenv("TAILSCALE_HOSTNAME")
	TAILSCALE_PORT__S := os.Getenv("TAILSCALE_PORT")
	TAILSCALE_PORT__64, err := strconv.ParseInt(TAILSCALE_PORT__S, 10, 32)
	if err != nil {
		return errors.New("TAILSCALE_PORT: " + err.Error())
	}
	TAILSCALE_PORT := int32(TAILSCALE_PORT__64)
	VOICEVOX_ENDPOINT := os.Getenv("VOICEVOX_ENDPOINT")
	env = environment{
		ANTHROPIC_API_KEY:        ANTHROPIC_API_KEY,
		BERTVITS2_ENDPOINT:       BERTVITS2_ENDPOINT,
		COHERE_API_KEY:           COHERE_API_KEY,
		DB_FILE_PATH:             DB_FILE_PATH,
		HOST:                     HOST,
		OPENAI_API_KEY:           OPENAI_API_KEY,
		PORT:                     PORT,
		STYLEBERTVIT2_ENDPOINT:   STYLEBERTVIT2_ENDPOINT,
		TAILSCALE_ENABLED:        TAILSCALE_ENABLED,
		TAILSCALE_ENABLED_TLS:    TAILSCALE_ENABLED_TLS,
		TAILSCALE_FUNNEL_ENABLED: TAILSCALE_FUNNEL_ENABLED,
		TAILSCALE_HOSTNAME:       TAILSCALE_HOSTNAME,
		TAILSCALE_PORT:           TAILSCALE_PORT,
		VOICEVOX_ENDPOINT:        VOICEVOX_ENDPOINT,
	}
	return err
}

type getterInterface interface {
	ANTHROPIC_API_KEY() string
	BERTVITS2_ENDPOINT() string
	COHERE_API_KEY() string
	DB_FILE_PATH() string
	HOST() string
	OPENAI_API_KEY() string
	PORT() int32
	STYLEBERTVIT2_ENDPOINT() string
	TAILSCALE_ENABLED() bool
	TAILSCALE_ENABLED_TLS() bool
	TAILSCALE_FUNNEL_ENABLED() bool
	TAILSCALE_HOSTNAME() string
	TAILSCALE_PORT() int32
	VOICEVOX_ENDPOINT() string
}
type getter struct {
	getterInterface
}

// Get returns a getter.
// getter is a struct for retrieving a value.
func Get() getter {
	return getter{}
}

type setterInterface interface {
	ANTHROPIC_API_KEY() string
	BERTVITS2_ENDPOINT() string
	COHERE_API_KEY() string
	DB_FILE_PATH() string
	HOST() string
	OPENAI_API_KEY() string
	PORT() int32
	STYLEBERTVIT2_ENDPOINT() string
	TAILSCALE_ENABLED() bool
	TAILSCALE_ENABLED_TLS() bool
	TAILSCALE_FUNNEL_ENABLED() bool
	TAILSCALE_HOSTNAME() string
	TAILSCALE_PORT() int32
	VOICEVOX_ENDPOINT() string
}
type setter struct {
	setterInterface
}

// Set returns a setter.
// setter is a struct for inserting a value.
func Set() setter {
	return setter{}
}

// Reset will reset the env variable.
func Reset() {
	env = environment{}
	return
}
