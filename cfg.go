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
	"strings"
	"time"
	"unicode"
)

type Config struct {
	Files []File
}

type File struct {
	Name string
	Path string
}

type CustomType interface {
	Unmarshal(value string) error
}

type FileSize uint64

func (fs *FileSize) Unmarshal(value string) error {
	size := FileSize(0)

	if len(value) == 0 {
		fs = &size
		return nil
	}

	value = strings.ToLower(value)
	last := len(value) - 1

	mp := uint64(1)

	if value[last] == 'b' {
		switch value[last-1] {
		case 't':
			mp = mp << 40
			value = strings.TrimSpace(value[:last-1])
		case 'g':
			mp = mp << 30
			value = strings.TrimSpace(value[:last-1])
		case 'm':
			mp = mp << 20
			value = strings.TrimSpace(value[:last-1])
		case 'k':
			mp = mp << 10
			value = strings.TrimSpace(value[:last-1])
		default:
			value = strings.TrimSpace(value[:last])
		}
	}

	ps, err := strconv.ParseUint(value, 10, 64)

	if err != nil {
		return err
	}

	*fs = FileSize(uint64(ps) * mp)
	return nil
}

func (c *Config) AddConfig(path, name string) {
	f := File{
		Path: path,
		Name: name,
	}

	c.Files = append(c.Files, f)
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

func LoadConfig(file string, dest interface{}) error {
	f, err := os.Open(file)

	if err != nil {
		return err
	}

	defer f.Close()

	err = parse(f, dest)

	if err != nil {
		return err
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
			continue
		}

		if line[0] == '#' {
			continue
		}

		kv := bytes.SplitN(line, []byte("="), 2)

		if len(kv) < 2 {
			continue
		}

		key := string(bytes.TrimRightFunc(kv[0], unicode.IsSpace))
		value := string(bytes.TrimSpace(bytes.TrimRight(kv[1], "\r\n")))
		kvmap[key] = value
	}

	err := searchFields(kvmap, dest)

	if err != nil {
		return err
	}

	return nil
}

const (
	tagCfg     = "cfg"
	tagDefault = "default"
)

func searchFields(kv map[string]string, dest interface{}) error {
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
		if el.Field(i).Kind() == reflect.Struct {
			searchFields(kv, el.Field(i).Addr().Interface())
			continue
		}
		if el.Field(i).CanSet() {
			sKey := el.Type().Field(i).Tag.Get(tagCfg)

			if sKey == "-" {
				continue
			}

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
				err := setField(el.Field(i), value)

				if err != nil {
					return err
				}

				delete(fieldDefaults, sKey)
			}
		}
	}

	for _, v := range fieldDefaults {
		err := setField(v.field, v.def)

		if err != nil {
			return err
		}
	}

	return nil
}

func setField(field reflect.Value, value string) error {
	customType := reflect.TypeOf((*CustomType)(nil)).Elem()

	if reflect.PtrTo(field.Type()).Implements(customType) {
		if c, ok := field.Addr().Interface().(CustomType); ok {
			err := c.Unmarshal(value)

			if err != nil {
				return err
			}
		}
		return nil
	}

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
