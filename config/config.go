package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Setting contains string only command line arguments in Key=>Next form.
type Setting struct {
	Key          string
	Description  string
	Value        string
	DefaultValue string
}

// Settings stores all the command line arguments with values.
type Settings map[string]*Setting

// Creates a new Settings by combining all Settings arguments, with increasing priority left to right.
func mergeSettings(settings ...Settings) Settings {
	combined := make(Settings)

	for _, s := range settings {
		for k := range s {
			combined[strings.ToLower(k)] = s[k]
		}
	}

	return combined
}

// LoadSettings loads the settings from commandline and the required config file.
func LoadSettings(args []string) (Settings, error) {
	// Start with command line to find the config file name
	clArgs, clErr := ParseCommandLine(args)
	if clErr != nil {
		return nil, clErr
	}

	cfName, cfPresent := clArgs.GetString("config.filename")
	if !cfPresent {
		return nil, fmt.Errorf("config.filename was not passed on command line")
	}

	// Load the config file
	cfArgs, cfErr := ParseConfigFile(cfName)
	if cfErr != nil {
		return nil, cfErr
	}

	combined := mergeSettings(cfArgs, clArgs)

	// What do we end up with?
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Debug(fmt.Sprintf("Current log level is %v\n", logrus.GetLevel()))
	for k, v := range combined {
		logrus.WithFields(logrus.Fields{"Key": k, "Next": v}).Debug("LoadSettings")
	}

	return combined, nil
}

// ParseConfigBuffer reads config from a generic Reader.
func ParseConfigBuffer(in io.Reader, format string) (Settings, error) {
	settings := make(Settings)

	viper.SetConfigType(format)
	err := viper.ReadConfig(in)
	if err == nil {
		for _, k := range viper.AllKeys() {
			m := viper.GetStringSlice(k)
			fmt.Printf("m:%q\n", m)

			v := viper.GetString(k)

			settings[k] = &Setting{Key: k, Value: v}
		}
	}

	return settings, err
}

// ParseConfigFile reads the file using Viper.
func ParseConfigFile(configFileName string) (Settings, error) {
	components := strings.Split(configFileName, ".")
	format := components[len(components)-1]

	settings := make(Settings)

	file, err := os.Open(configFileName)
	if err == nil {
		configBuffer := make([]byte, 10240)
		sz, err := file.Read(configBuffer)
		if err == nil {
			settings, err = ParseConfigBuffer(bytes.NewBuffer(configBuffer[:sz]), format)
			if err != nil {
				return nil, err
			}
		}
	}

	return settings, err
}

// ParseCommandLine takes an array of string and turns them into keyvalue pairs.
func ParseCommandLine(arguments []string) (Settings, error) {
	args := make(Settings)
	lookingForValue := false
	currentArg := &Setting{}
	for _, val := range arguments {
		if lookingForValue {
			currentArg.Value = val
			args[currentArg.Key] = currentArg
			lookingForValue = false
		} else {
			if strings.HasPrefix(val, "--") || strings.HasPrefix(val, "-") {
				currentArg = new(Setting)
				val = strings.TrimPrefix(strings.TrimPrefix(val, "-"), "-")
				switch s := strings.Split(val, "="); len(s) {
				case 1:
					currentArg.Key = val
					lookingForValue = true
				case 2:
					currentArg.Key = s[0]
					currentArg.Value = s[1]
					args[currentArg.Key] = currentArg
				default:
					mess := fmt.Sprintf("too many '='s in argument %v", currentArg.Key)
					return nil, errors.New(mess)
				}
			}
		}
	}
	if lookingForValue {
		args[currentArg.Key] = nil
	}

	return args, nil
}

// RequiredSettingsPresent Foo.
func (settings Settings) RequiredSettingsPresent(requiredSettings Settings) error {
	missingFlags := []string{}

	for flag := range requiredSettings {
		_, present := settings[flag]
		if !present {
			missingFlags = append(missingFlags, flag)
		}
	}

	if len(missingFlags) > 0 {
		return fmt.Errorf("missing required settings: %s", strings.Join(missingFlags, ","))
	}
	return nil
}

// FilterNamespaces returns a copy of settings belonging to given namespace or no namespace.
func (settings Settings) FilterNamespaces(nsKeep string) Settings {
	remaining := make(Settings)

	prefix := fmt.Sprintf("%s.", nsKeep)

	for k, v := range settings {
		if strings.HasPrefix(k, prefix) || !strings.Contains(k, ".") {
			remaining[k] = v
		}
	}

	return remaining
}

// GetString returns value of given commandline flag.
func (settings Settings) GetString(name string) (string, bool) {
	flag, present := settings[strings.ToLower(name)]

	val := ""
	if present && flag != nil {
		val = flag.Value
	}

	logrus.WithFields(logrus.Fields{"Key": name, "Next": val, "Present": present}).Debug("GetString")

	return val, present
}
