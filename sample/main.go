package main

import (
	"fmt"
	"os"
	"strings"

	"git.hoogi.eu/cfg"
)

type Settings struct {
	ServerPort int          `cfg:"server_port" default:"8080"`
	Filesize   cfg.FileSize `cfg:"filesize"`
	Log
}

type Log struct {
	File  string   `cfg:"log_file"`
	Level LogLevel `cfg:"log_level"`
}

type LogLevel int

const (
	Info = iota
	Debug
)

func (lm *LogLevel) Unmarshal(value string) error {
	if strings.ToLower(value) == "info" {
		*lm = LogLevel(Info)
		return nil
	} else if strings.ToLower(value) == "debug" {
		*lm = LogLevel(Debug)
		return nil
	}
	return fmt.Errorf("unexpected config value '%s' for log level", value)
}

func main() {
	c := cfg.Config{}
	c.AddConfig(".", "myconfig.conf")

	settings := new(Settings)
	err := c.MergeConfigsInto(settings)

	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	fmt.Println(settings.ServerPort) // 8080
	fmt.Println(settings.Filesize)   // 10485760
	fmt.Println(settings.Log.File)   // /var/log/my.log
	fmt.Println(settings.Log.Level)  // 1
}
