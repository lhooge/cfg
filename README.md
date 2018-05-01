# Simple config reading for Go using reflection

Usage
-----

Sample config file:

    # A comment
    filesize = 10MB
    log_level = Debug
    log_file = /var/log/my.log

Quotes are not trimmed.
Whitespaces before and after the value are trimmed.

Unmarshal config into struct

    type Settings struct {
        ServerPort int     `cfg:"server_port" default:"8080"`	// config value; if not found in config default is used
        Filesize cfg.Filesize `cfg:"filesize"`			// returns bytes of specified filesize 
        Log
    }

    type Log struct {
        File    string   `cfg:"log_file"`	// config value without default; if not found an error is returned
        Level   LogLevel `cfg:"log_level"`	// custom LogLevel type unmarshals the value "info" or "debug" into type LogLevel
    }

    type LogLevel int				// custom int type which holds the log level

    const (
        Info = iota				// iota value for log levels
        Debug
    )

    func (lm *LogLevel) Unmarshal(value string) error {		// the custom unmarshaler for the log level
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
        c := cfg.Config{}				// create a new config which holds the an array of files
        c.AddConfig("/etc", myconfig.conf)		// convenient method for adding a file
        
        settings := new(Settings)		
        def, err := c.MergeConfigsInto(settings)	// merges the config values into the setting struct
	
        if err != nil {
             panic(err)
        }
	
	for k, v := range def {				// applied defaults
	    fmt.Printf("using default value %s for key %s\n", v.Value, k)
	}

        fmt.Println(settings.ServerPort)
        fmt.Println(settings.Filesize)
        fmt.Println(settings.Log.File)
        fmt.Println(settings.Log.Level)
    }
