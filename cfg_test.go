package cfg

import (
	"testing"
	"time"
)

func TestStandardConfig(t *testing.T) {
	type config struct {
		SessionName  string `cfg:"session_name"`
		FileLocation string `cfg:"file_location" default:"/dev/null"`
		Address      string

		Port int `cfg:"port" default:"2000"`
		Size int `default:"30"`

		SSL     bool `cfg:"ssl"`
		Verbose bool `cfg:"verbose" default:"yes"`

		SessionTimeout time.Duration `cfg:"session_timeout"`
	}

	c := config{}

	AddConfig("./testcfg", "config.conf")

	err := cfg.MergeConfigsInto(&c)

	if err != nil {
		t.Error(err)
	}

	if c.SessionName != "the-session-name" {
		t.Errorf("invalid value %s", c.SessionName)
	}

	if c.FileLocation != "/dev/null" {
		t.Errorf("FileLocation expected to be /dev/null but was %s", c.FileLocation)
	}

	if c.Address != "127.0.0.1" {
		t.Errorf("Address expected to be 127.0.0.1 but was %s", c.Address)
	}

	if c.Port != 8080 {
		t.Errorf("Port expected to be 8080 but was %d", c.Port)
	}

	if c.Size != 30 {
		t.Errorf("Size expected to be 30 but was %d", c.Size)
	}

	if !c.SSL {
		t.Errorf("SSL expected to be true but was %t", c.SSL)
	}

	if !c.Verbose {
		t.Errorf("Verbose expected to be true but was %t", c.Verbose)
	}

	expDuration, _ := time.ParseDuration("10m")
	if c.SessionTimeout != expDuration {
		t.Errorf("SessionTimeout expected to be  but was %v", c.SessionTimeout)
	}
}

func TestInnerStruct(t *testing.T) {
	type Config struct {
		Server struct {
			Address string `cfg:"server_address"`
			Port    int    `cfg:"server_port"`
		}
	}

	AddConfig("./testcfg", "config.conf")

	c := new(Config)
	err := cfg.MergeConfigsInto(c)

	if err != nil {
		t.Error(err)
	}

	if c.Server.Address != "localhost" {
		t.Errorf("server.Address expected to be localhost but was %s", c.Server.Address)
	}
	if c.Server.Port != 42 {
		t.Errorf("server.Port expected to be 42 but was %d", c.Server.Port)
	}
}
