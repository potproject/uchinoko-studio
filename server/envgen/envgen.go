// Code generated by github.com/potproject/goenvgen, DO NOT EDIT.

package envgen

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

func (g getter) DB_FILE_PATH() string {
	return env.DB_FILE_PATH
}
func (s setter) DB_FILE_PATH(value string) {
	env.DB_FILE_PATH = value
	return
}
func (g getter) ELEVENLABS_API_KEY() string {
	return env.ELEVENLABS_API_KEY
}
func (s setter) ELEVENLABS_API_KEY(value string) {
	env.ELEVENLABS_API_KEY = value
	return
}
func (g getter) ELEVENLABS_OUTPUT_FORMAT() string {
	return env.ELEVENLABS_OUTPUT_FORMAT
}
func (s setter) ELEVENLABS_OUTPUT_FORMAT(value string) {
	env.ELEVENLABS_OUTPUT_FORMAT = value
	return
}
func (g getter) ELEVENLABS_VOICEID() string {
	return env.ELEVENLABS_VOICEID
}
func (s setter) ELEVENLABS_VOICEID(value string) {
	env.ELEVENLABS_VOICEID = value
	return
}
func (g getter) OPENAI_API_KEY() string {
	return env.OPENAI_API_KEY
}
func (s setter) OPENAI_API_KEY(value string) {
	env.OPENAI_API_KEY = value
	return
}
func (g getter) OPENAI_CHAT_MODEL() string {
	return env.OPENAI_CHAT_MODEL
}
func (s setter) OPENAI_CHAT_MODEL(value string) {
	env.OPENAI_CHAT_MODEL = value
	return
}
func (g getter) OPENAI_ORG_ID() string {
	return env.OPENAI_ORG_ID
}
func (s setter) OPENAI_ORG_ID(value string) {
	env.OPENAI_ORG_ID = value
	return
}
func (g getter) PORT() int32 {
	return env.PORT
}
func (s setter) PORT(value int32) {
	env.PORT = value
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
func (g getter) VOICEVOX_SPEAKER() int64 {
	return env.VOICEVOX_SPEAKER
}
func (s setter) VOICEVOX_SPEAKER(value int64) {
	env.VOICEVOX_SPEAKER = value
	return
}

type environment struct {
	DB_FILE_PATH             string
	ELEVENLABS_API_KEY       string
	ELEVENLABS_OUTPUT_FORMAT string
	ELEVENLABS_VOICEID       string
	OPENAI_API_KEY           string
	OPENAI_CHAT_MODEL        string
	OPENAI_ORG_ID            string
	PORT                     int32
	TAILSCALE_ENABLED        bool
	TAILSCALE_ENABLED_TLS    bool
	TAILSCALE_HOSTNAME       string
	TAILSCALE_PORT           int32
	VOICEVOX_ENDPOINT        string
	VOICEVOX_SPEAKER         int64
}

var env environment

// Load reads the environment variables and stores them in the env variable.
// If the type conversion fails, it returns error.
func Load() error {
	var err error
	DB_FILE_PATH := os.Getenv("DB_FILE_PATH")
	ELEVENLABS_API_KEY := os.Getenv("ELEVENLABS_API_KEY")
	ELEVENLABS_OUTPUT_FORMAT := os.Getenv("ELEVENLABS_OUTPUT_FORMAT")
	ELEVENLABS_VOICEID := os.Getenv("ELEVENLABS_VOICEID")
	OPENAI_API_KEY := os.Getenv("OPENAI_API_KEY")
	OPENAI_CHAT_MODEL := os.Getenv("OPENAI_CHAT_MODEL")
	OPENAI_ORG_ID := os.Getenv("OPENAI_ORG_ID")
	PORT__S := os.Getenv("PORT")
	PORT__64, err := strconv.ParseInt(PORT__S, 10, 32)
	if err != nil {
		return errors.New("PORT: " + err.Error())
	}
	PORT := int32(PORT__64)
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
	TAILSCALE_HOSTNAME := os.Getenv("TAILSCALE_HOSTNAME")
	TAILSCALE_PORT__S := os.Getenv("TAILSCALE_PORT")
	TAILSCALE_PORT__64, err := strconv.ParseInt(TAILSCALE_PORT__S, 10, 32)
	if err != nil {
		return errors.New("TAILSCALE_PORT: " + err.Error())
	}
	TAILSCALE_PORT := int32(TAILSCALE_PORT__64)
	VOICEVOX_ENDPOINT := os.Getenv("VOICEVOX_ENDPOINT")
	VOICEVOX_SPEAKER__S := os.Getenv("VOICEVOX_SPEAKER")
	VOICEVOX_SPEAKER__64, err := strconv.ParseInt(VOICEVOX_SPEAKER__S, 10, 64)
	if err != nil {
		return errors.New("VOICEVOX_SPEAKER: " + err.Error())
	}
	VOICEVOX_SPEAKER := int64(VOICEVOX_SPEAKER__64)
	env = environment{
		DB_FILE_PATH:             DB_FILE_PATH,
		ELEVENLABS_API_KEY:       ELEVENLABS_API_KEY,
		ELEVENLABS_OUTPUT_FORMAT: ELEVENLABS_OUTPUT_FORMAT,
		ELEVENLABS_VOICEID:       ELEVENLABS_VOICEID,
		OPENAI_API_KEY:           OPENAI_API_KEY,
		OPENAI_CHAT_MODEL:        OPENAI_CHAT_MODEL,
		OPENAI_ORG_ID:            OPENAI_ORG_ID,
		PORT:                     PORT,
		TAILSCALE_ENABLED:        TAILSCALE_ENABLED,
		TAILSCALE_ENABLED_TLS:    TAILSCALE_ENABLED_TLS,
		TAILSCALE_HOSTNAME:       TAILSCALE_HOSTNAME,
		TAILSCALE_PORT:           TAILSCALE_PORT,
		VOICEVOX_ENDPOINT:        VOICEVOX_ENDPOINT,
		VOICEVOX_SPEAKER:         VOICEVOX_SPEAKER,
	}
	return err
}

type getterInterface interface {
	DB_FILE_PATH() string
	ELEVENLABS_API_KEY() string
	ELEVENLABS_OUTPUT_FORMAT() string
	ELEVENLABS_VOICEID() string
	OPENAI_API_KEY() string
	OPENAI_CHAT_MODEL() string
	OPENAI_ORG_ID() string
	PORT() int32
	TAILSCALE_ENABLED() bool
	TAILSCALE_ENABLED_TLS() bool
	TAILSCALE_HOSTNAME() string
	TAILSCALE_PORT() int32
	VOICEVOX_ENDPOINT() string
	VOICEVOX_SPEAKER() int64
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
	DB_FILE_PATH() string
	ELEVENLABS_API_KEY() string
	ELEVENLABS_OUTPUT_FORMAT() string
	ELEVENLABS_VOICEID() string
	OPENAI_API_KEY() string
	OPENAI_CHAT_MODEL() string
	OPENAI_ORG_ID() string
	PORT() int32
	TAILSCALE_ENABLED() bool
	TAILSCALE_ENABLED_TLS() bool
	TAILSCALE_HOSTNAME() string
	TAILSCALE_PORT() int32
	VOICEVOX_ENDPOINT() string
	VOICEVOX_SPEAKER() int64
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
