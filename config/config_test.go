package config

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGood(t *testing.T) {
	goodArgs := [][]string{
		{"--name", "me"},
		{"--age", "1000000", "--height=73"},
		{"--foo=f"},
		{"Just a thing", "and another"},
		{"yet another", "--age=2"},
		{"--dangling"},
		{},
	}

	for _, goodArgs := range goodArgs {
		_, err := ParseCommandLine(goodArgs)
		assert.True(t, err == nil, "Should be no parsing errors")
	}
}

func TestParseBad(t *testing.T) {
	badArgs := []string{"--foo==bar"}
	_, clErr := ParseCommandLine(badArgs)
	assert.True(t, clErr != nil, "Should reject extra ='s")
}

func TestAlternateForms(t *testing.T) {
	argNames := [][]string{
		{"--name", "me"},
		{"--name=me"},
		{"filler", "--name", "me"},
	}

	for _, argName := range argNames {
		flags, _ := ParseCommandLine(argName)
		nameValue, namePresent := flags.GetString("name")

		assert.True(t, namePresent, "Name is a flag")
		assert.Equal(t, "me", nameValue, "Name should be me")
	}
}

func TestValueNotProvided(t *testing.T) {
	flags, _ := ParseCommandLine([]string{"--name=me", "--help"})
	helpValue, helpPresent := flags.GetString("help")
	assert.True(t, helpPresent, "", "flag present without value")
	assert.Equal(t, "", helpValue, "flag present without value")
}

func TestNotProvided(t *testing.T) {
	flags, _ := ParseCommandLine([]string{"--name=me"})
	signValue, signPresent := flags.GetString("sign")

	assert.True(t, !signPresent, "Sign was not provided. FYI - Snake")
	assert.Equal(t, "", signValue, "Empty string is returns")
}

func TestConfigSimple(t *testing.T) {
	const configText = `[config]
						name = "Marge"`

	flags, _ := ParseConfigBuffer(bytes.NewBuffer([]byte(configText)), "toml")

	name, present := flags.GetString("config.name")
	assert.True(t, present, "Name is present in config")
	assert.Equal(t, "Marge", name, "Name is set to 'Marge'")
}

func TestConfigNotProvided(t *testing.T) {
	const configText = `[config]
						name = "Marge"`

	flags, _ := ParseConfigBuffer(bytes.NewBuffer([]byte(configText)), "toml")

	name, present := flags.GetString("config.sign")
	assert.True(t, !present, "Name is present in config")
	assert.Equal(t, "", name, "Name is set to 'Marge'")
}

func TestConfigFile(t *testing.T) {
	flags, _ := ParseConfigFile("config_test.toml")

	name, present := flags.GetString("config.sign")
	assert.True(t, present, "Sign is present in config")
	assert.Equal(t, "Snake", name, "My Chinese Zodiac sign is 'Snake'")
}

func TestLoadSettingsMergePresent(t *testing.T) {
	commandLine := []string{"--config.filename", "config_test.toml"}

	settings, _ := LoadSettings(commandLine)

	cf, present := settings.GetString("config.filename")
	assert.True(t, present, "config.filename is present")
	assert.Equal(t, "config_test.toml", cf, "config file name is config_test.toml")

	cf, present = settings.GetString("config.sign")
	assert.True(t, present, "config.sign is present")
	assert.Equal(t, "Snake", cf, "config sign is Snake")
}

func TestLoadSettingsMergeFileMissing(t *testing.T) {
	commandLine := []string{"--config.filename", "bob.toml"}

	_, err := LoadSettings(commandLine)

	assert.True(t, err != nil, "error due to missing filename")
}

func TestLoadSettingsMergeFileNameMissing(t *testing.T) {
	commandLine := []string{"--config.not_file_name", "config_test.toml"}

	_, err := LoadSettings(commandLine)

	assert.True(t, err != nil, "error due to missing filename")
}

func TestLoadSettingsMergeMissing(t *testing.T) {
	commandLine := []string{"--config.filename", "config_test.toml", "--config.sign", "Ox"}

	settings, _ := LoadSettings(commandLine)

	cf, present := settings.GetString("swallow.airspeed")
	assert.True(t, !present, "swallow.airspeed is present")
	assert.Equal(t, "", cf, "swallow.airspeed is empty")
}

func TestLoadSettingsParseError(t *testing.T) {
	commandLine := []string{"--config.filename", "config_test.toml", "--config.sign==Ox"}

	_, err := LoadSettings(commandLine)

	assert.True(t, err != nil, "command line args are invalid")
}

func TestLoadSettingsOverride(t *testing.T) {
	commandLine := []string{"--config.filename", "config_test.toml", "--config.sign", "Ox"}

	settings, _ := LoadSettings(commandLine)

	cf, present := settings.GetString("config.sign")
	assert.True(t, present, "config.sign is present")
	assert.Equal(t, "Ox", cf, "config sign is overridden in commandline to Ox")
}

func TestLoadFilterNamespaces(t *testing.T) {
	commandLine := []string{
		"--foo=bar",
		"--favorite.cheese", "gouda",
		"--config.filename", "config_test.toml",
		"--config.sign", "Ox",
	}

	settings, _ := LoadSettings(commandLine)
	settings = settings.FilterNamespaces("config")

	cf, present := settings.GetString("config.sign")
	assert.True(t, present, "config.sign is present")
	assert.Equal(t, "Ox", cf, "config sign is overridden in commandline to Ox")

	cf, present = settings.GetString("favorite.cheese")
	assert.True(t, !present, "favorites namespace should be removed now")
	assert.Equal(t, "", cf, "favorites were removed so values should all be empty")

	cf, present = settings.GetString("foo")
	assert.True(t, present, "foo is not namespaced so should still be present")
	assert.Equal(t, "bar", cf, "foo is still present")
}

func TestRequiredSettingsPresent(t *testing.T) {
	requiredSettings := map[string]*Setting{"foo": {Key: "foo"}}

	commandLine := []string{"--foo", "ack"}
	settings, _ := ParseCommandLine(commandLine)

	err := settings.RequiredSettingsPresent(requiredSettings)
	assert.True(t, err == nil, "Should be nothing missing")
}

func TestRequiredSettingsNotPresent(t *testing.T) {
	requiredSettings := map[string]*Setting{"sign": {Key: "Snake"}}

	commandLine := []string{"--foo", "ack"}
	settings, _ := ParseCommandLine(commandLine)

	err := settings.RequiredSettingsPresent(requiredSettings)
	assert.True(t, err != nil, "Should be nothing missing")
}

func TestConfigFileArray(t *testing.T) {
	settings, _ := ParseConfigFile("config_test.toml")

	bootnodes, present := settings.GetString("p2p.bootnodes")
	assert.True(t, present, "Bootnodes are in the file")

	fmt.Printf("bootnodes:%v\n", bootnodes)
}
