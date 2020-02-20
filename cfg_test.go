package cfg_test

import (
	"testing"
	"time"

	"git.hoogi.eu/snafu/cfg"
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

	def, err := c.MergeConfigsInto(s)

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

	if len(def) != 3 {
		t.Error("expected three entries in applied default values")
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

	_, err := c.MergeConfigsInto(s)

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

func TestHumanizeFilesize(t *testing.T) {
	var fileSizes = []struct {
		fs       cfg.FileSize
		expected string
	}{
		{cfg.FileSize(0), "0"},
		{cfg.FileSize(6680), "6.6 KB"},
		{cfg.FileSize(4314), "4.3 KB"},
		{cfg.FileSize(237797290), "226.8 MB"},
		{cfg.FileSize(4096), "4.0 KB"},
	}

	for _, v := range fileSizes {
		actual := v.fs.HumanReadable()

		if actual != v.expected {
			t.Errorf("HumanReadable(%d): expected %s, actual %s", uint64(v.fs), v.expected, actual)
		}
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
	*lm = m
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

	_, err := c.MergeConfigsInto(s)

	if err != nil {
		t.Error(err)
	}

	if s.Custom != username {
		t.Errorf("s.Custom expected to be username but was %s", s.Custom)
	}
}

func TestFileSizes(t *testing.T) {
	type settings struct {
		Byte     cfg.FileSize `cfg:"file_size_byte"`
		Kilobyte cfg.FileSize `cfg:"file_size_kilobyte"`
		Megabyte cfg.FileSize `cfg:"file_size_megabyte"`
		Gigabyte cfg.FileSize `cfg:"file_size_gigabyte"`
		Terabyte cfg.FileSize `cfg:"file_size_terabyte"`
		Empty    cfg.FileSize `cfg:"file_size_empty"`
	}

	c := addConfig("./testcfg", "config.conf")

	s := new(settings)

	_, err := c.MergeConfigsInto(s)

	if err != nil {
		t.Error(err)
	}

	if s.Byte != 5 {
		t.Errorf("s.Byte expected to be %d bytes but was %d", 5, s.Byte)
	}
	if s.Kilobyte != 1<<10 {
		t.Errorf("s.Kilobyte expected to be %d bytes but was %d", 1<<10, s.Kilobyte)
	}
	if s.Megabyte != 1<<20 {
		t.Errorf("s.Megabyte expected to be %d but was %d", 1<<20, s.Megabyte)
	}
	if s.Gigabyte != 1<<30 {
		t.Errorf("s.Gigabyte expected to be %d but was %d", 1<<30, s.Gigabyte)
	}
	if s.Terabyte != 1<<40 {
		t.Errorf("s.Terabyte expected to be %d but was %d", 1<<40, s.Terabyte)
	}
	if s.Empty != 0 {
		t.Errorf("s.Empty expected to be 0 but was %d", s.Empty)
	}

}

func addConfig(path, filename string) cfg.ConfigFiles {
	cfg := cfg.ConfigFiles{
		Files: make([]cfg.File, 0, 1),
	}

	cfg.AddConfig(path, filename, true)

	return cfg
}
