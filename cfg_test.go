package cfg

import (
	"testing"
	"time"
)

func TestStandardConfig(t *testing.T) {
	type settings struct {
		SessionName  string `cfg:"session_name"`
		FileLocation string `cfg:"file_location" default:"/dev/null"`
		Address      string

		Port int `cfg:"port" default:"2000"`
		Size int `default:"30"`

		SSL     bool `cfg:"ssl"`
		Verbose bool `cfg:"verbose" default:"yes"`

		SessionTimeout time.Duration `cfg:"session_timeout"`
	}

	c := addConfig("./testcfg", "config.conf")

	s := new(settings)

	err := c.MergeConfigsInto(s)

	if err != nil {
		t.Error(err)
	}

	if s.SessionName != "the-session-name" {
		t.Errorf("invalid value %s", s.SessionName)
	}

	if s.FileLocation != "/dev/null" {
		t.Errorf("FileLocation expected to be /dev/null but was %s", s.FileLocation)
	}

	if s.Address != "127.0.0.1" {
		t.Errorf("Address expected to be 127.0.0.1 but was %s", s.Address)
	}

	if s.Port != 8080 {
		t.Errorf("Port expected to be 8080 but was %d", s.Port)
	}

	if s.Size != 30 {
		t.Errorf("Size expected to be 30 but was %d", s.Size)
	}

	if !s.SSL {
		t.Errorf("SSL expected to be true but was %t", s.SSL)
	}

	if !s.Verbose {
		t.Errorf("Verbose expected to be true but was %t", s.Verbose)
	}

	expDuration, _ := time.ParseDuration("10m")
	if s.SessionTimeout != expDuration {
		t.Errorf("SessionTimeout expected to be  but was %v", s.SessionTimeout)
	}
}

func TestInnerStruct(t *testing.T) {
	type settings struct {
		Server struct {
			Address string `cfg:"server_address"`
			Port    int    `cfg:"server_port"`
		}
	}

	c := addConfig("./testcfg", "config.conf")

	s := new(settings)

	err := c.MergeConfigsInto(s)

	if err != nil {
		t.Error(err)
	}

	if s.Server.Address != "localhost" {
		t.Errorf("server.Address expected to be localhost but was %s", s.Server.Address)
	}
	if s.Server.Port != 42 {
		t.Errorf("server.Port expected to be 42 but was %d", s.Server.Port)
	}
}

func TestArrayConfig(t *testing.T) {
	type settings struct {
		GroupList []string `cfg:"group_list"`
	}

	c := addConfig("./testcfg", "config.conf")

	s := new(settings)

	err := c.MergeConfigsInto(s)

	if err != nil {
		t.Error(err)
	}
}

type loginMethod int

const (
	mail = iota
	username
)

func (lm *loginMethod) Unmarshal(value string) error {
	m := loginMethod(username)
	if value == "mail" {
		m = loginMethod(mail)
	}
	lm = &m
	return nil
}

func (lm loginMethod) String() string {
	if lm == mail {
		return "mail"
	} else {
		return "username"
	}
}

func TestCustomType(t *testing.T) {
	type settings struct {
		Custom loginMethod `cfg:"login_method"`
	}

	c := addConfig("./testcfg", "config.conf")

	s := new(settings)

	err := c.MergeConfigsInto(s)

	if err != nil {
		t.Error(err)
	}

	if s.Custom != mail {
		t.Errorf("s.Custom expected to be mail but was %s", s.Custom)
	}
}

func addConfig(path, filename string) Config {
	cfg := Config{
		Files: make([]File, 0, 1),
	}

	cfg.AddConfig(path, filename)

	return cfg
}
