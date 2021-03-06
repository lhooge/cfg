package cfg

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	tagCfg     = "cfg"
	tagDefault = "default"
)

// ConfigFiles represents multiple file containing the config keys and values
type ConfigFiles struct {
	Files []File
}

// File represents a file
// Required if an error should be thrown if file is absent
type File struct {
	Name     string
	Path     string
	Required bool
}

// Default represents a default value for a field
type Default struct {
	Value string
	field reflect.Value
}

// CustomType can be implemented to unmarshal in a custom format
type CustomType interface {
	Unmarshal(value string) error
}

// FileSize implements Unmarshal for parsing a file size config value.
// e.g. 10MB
type FileSize uint64

const (
	B = 1 << (iota * 10)
	KB
	MB
	GB
	TB
)

var sizes = []string{"B", "KB", "MB", "GB", "TB"}

//HumanReadable returns a human readable form of the filesize e.g 12.5 MB, 1.0 GB
func (fs FileSize) HumanReadable() string {
	if fs == 0 {
		return "0"
	}

	exp := math.Floor(math.Log(float64(fs)) / math.Log(1024))

	if exp > 4 {
		exp = 4
	}

	s := sizes[int(exp)]

	if exp == 0 {
		return fmt.Sprintf("%d %s", fs, s)
	}

	val := float64(fs) / float64(math.Pow(1024, exp))

	return fmt.Sprintf("%.1f %s", math.Ceil(float64(val)*10)/10, s)
}

func (fs *FileSize) Unmarshal(value string) error {
	size := FileSize(0)

	if len(value) == 0 {
		fs = &size
		return nil
	}

	value = strings.TrimSpace(strings.ToLower(value))
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

//AddConfig adds a config file
func (c *ConfigFiles) AddConfig(path, name string, required bool) {
	f := File{
		Path:     path,
		Name:     name,
		Required: required,
	}

	c.Files = append(c.Files, f)
}

//MergeConfigsInto merges multiple configs files into a struct
//returns the applied default values
func (c ConfigFiles) MergeConfigsInto(dest interface{}) (map[string]Default, error) {
	kvs := make(map[string]string)

	for _, v := range c.Files {
		f, err := os.Open(filepath.Join(v.Path, v.Name))

		if err != nil {
			if !v.Required && os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		defer f.Close()

		kv, err := parse(f, dest)

		if err != nil {
			return nil, err
		}

		for k, v := range kv {
			kvs[k] = v
		}
	}

	defaults := make(map[string]Default)
	err := setFields(kvs, defaults, dest)

	if err != nil {
		return nil, err
	}

	return defaults, nil
}

//LoadConfigInto loads a single config into struct
//returns the applied default values
func LoadConfigInto(file string, dest interface{}) (map[string]Default, error) {
	f, err := os.Open(file)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	kvs, err := parse(f, dest)

	if err != nil {
		return nil, err
	}

	defaults := make(map[string]Default)

	err = setFields(kvs, defaults, dest)

	if err != nil {
		return nil, err
	}

	return defaults, nil
}

func parse(file *os.File, dest interface{}) (map[string]string, error) {
	scanner := bufio.NewScanner(file)
	kvmap := make(map[string]string)

	for scanner.Scan() {
		line := scanner.Bytes()

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
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return kvmap, nil
}

func setFields(kv map[string]string, defaults map[string]Default, dest interface{}) error {
	v := reflect.ValueOf(dest)

	if v.Kind() != reflect.Ptr {
		return errors.New("struct must be a pointer")
	}

	el := v.Elem()

	for i := 0; i < el.NumField(); i++ {
		if el.Field(i).Kind() == reflect.Struct {
			err := setFields(kv, defaults, el.Field(i).Addr().Interface())
			if err != nil {
				return err
			}
			continue
		}
		if el.Field(i).CanSet() {
			sKey := el.Type().Field(i).Tag.Get(tagCfg)
			defValue := el.Type().Field(i).Tag.Get(tagDefault)

			if sKey == "-" {
				continue
			}

			if len(sKey) == 0 {
				sKey = el.Type().Field(i).Name
			}

			def := Default{}

			if len(defValue) > 0 {
				def = Default{
					Value: defValue,
					field: el.Field(i),
				}

				defaults[sKey] = def
			}

			value, ok := kv[sKey]

			if ok {
				err := setField(el.Field(i), value)

				if err != nil {
					if def != (Default{}) {
						//ignore error here if key has a default
						continue
					}
					return fmt.Errorf("error while setting value [%s] for key [%s] error %v", value, sKey, err)
				}

				delete(defaults, sKey)
			}
		}
	}
	for k, d := range defaults {
		err := setField(d.field, d.Value)
		if err != nil {
			return fmt.Errorf("error while setting default value [%s] for key [%s] error %v", d.Value, k, err)
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
