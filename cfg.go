package cfg

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"time"
	"unicode"
)

var cfg *Config

func init() {
	cfg = new(Config)
}

type Config struct {
	Files []File
}

type File struct {
	Name string
	Path string
}

func AddConfig(path, name string) {
	f := File{
		Path: path,
		Name: name,
	}

	cfg.Files = append(cfg.Files, f)
}

func (c Config) MergeConfigsInto(dest interface{}) error {
	for _, v := range c.Files {
		f, err := os.Open(filepath.Join(v.Path, v.Name))

		if err != nil {
			return err
		}

		defer f.Close()

		err = parse(f, dest)

		if err != nil {
			return err
		}
	}

	return nil
}

func parse(file *os.File, dest interface{}) error {
	reader := bufio.NewReader(file)
	kvmap := make(map[string]string)

	for {
		line, err := reader.ReadBytes('\n')

		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		line = bytes.TrimLeftFunc(line, unicode.IsSpace)

		if len(line) == 0 {
			break
		}

		if line[0] == '#' {
			break
		}

		kv := bytes.SplitN(line, []byte("="), 2)

		if len(kv) < 2 {
			break
		}

		key := string(bytes.TrimRightFunc(kv[0], unicode.IsSpace))
		value := string(bytes.TrimSpace(bytes.TrimRight(kv[1], "\r\n")))
		kvmap[key] = value
	}

	err := setField(kvmap, dest)

	if err != nil {
		return err
	}

	return nil
}

const (
	tagCfg     = "cfg"
	tagDefault = "default"
)

func setField(kv map[string]string, dest interface{}) error {
	v := reflect.ValueOf(dest)

	if v.Kind() != reflect.Ptr {
		return errors.New("struct must be a pointer")
	}

	type reflectDefaults struct {
		field reflect.Value
		def   string
	}

	fieldDefaults := make(map[string]reflectDefaults)

	el := v.Elem()

	for i := 0; i < el.NumField(); i++ {
		if el.Field(i).CanSet() {
			sKey := el.Type().Field(i).Tag.Get(tagCfg)

			def := reflectDefaults{
				field: el.Field(i),
				def:   el.Type().Field(i).Tag.Get(tagDefault),
			}

			if len(sKey) == 0 {
				sKey = el.Type().Field(i).Name
			}

			fieldDefaults[sKey] = def
			value, ok := kv[sKey]

			if ok {
				err := setType(el.Field(i), value)

				if err != nil {
					return err
				}

				delete(fieldDefaults, sKey)
			}
		}
	}

	for _, v := range fieldDefaults {
		err := setType(v.field, v.def)

		if err != nil {
			return err
		}
	}

	return nil
}

func setType(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int8, reflect.Int16, reflect.Int, reflect.Int64:
		d, err := time.ParseDuration(value)

		if err == nil {
			field.Set(reflect.ValueOf(d))
			return nil
		}

		iVal, err := strconv.ParseInt(value, 10, 64)

		if err != nil {
			return err
		}

		field.SetInt(int64(iVal))
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		iVal, err := strconv.ParseUint(value, 10, 64)

		if err != nil {
			return err
		}

		field.SetUint(uint64(iVal))
	case reflect.Bool:
		b := false

		if value == "yes" || value == "YES" || value == "Yes" {
			b = true
		} else if value == "no" || value == "NO" || value == "No" {
			b = false
		} else {
			var err error
			b, err = strconv.ParseBool(value)

			if err != nil {
				return err
			}
		}

		field.SetBool(b)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)

		if err != nil {
			return err
		}

		field.SetFloat(float64(f))
	}

	return nil
}
