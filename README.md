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
        //ServerPort config value; if key not found in config the default 8080 is used
        ServerPort int        `cfg:"server_port" default:"8080"`
        //Filesize returns bytes of specified filesize 
        Filesize cfg.Filesize `cfg:"filesize"`
        //Log the nested struct
        Log
    }

    type Log struct {
        //File config key without default value; if not found an error is returned
        File    string   `cfg:"log_file"`	
        //Level type LogLevel unmarshals the value "info" or "debug" into type LogLevel
        Level   LogLevel `cfg:"log_level"`	
    }

    // LogLevel custom int type which holds the log level
    type LogLevel int

    //Possible log levels
    const (    
        //Info logs everything with severity "Info"
        Info = iota
        //Debug logs everything with severity "Debug"
        Debug
    )

    // Unmarshal the custom unmarshaler for the log level
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
        // create a new ConfigFiles structure
        c := cfg.ConfigFiles{}
        // search for a config 'myconfig.conf' in folder '/etc' it fails if file is not found.
        c.AddConfig("/etc", myconfig.conf, true)	
        
        settings := new(Settings)		
        // merges the config values into the settings struct
        def, err := c.MergeConfigsInto(settings)	
	
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
