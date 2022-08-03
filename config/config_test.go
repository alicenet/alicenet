package config

import (
	"bytes"
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
		{}}

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
		{"filler", "--name", "me"}}

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

func TestLoadSettingsMergeFileMissing(t *testing.T) {

	commandLine := []string{"--config.filename", "bob.toml"}

	_, err := LoadSettings(commandLine)

	assert.True(t, err != nil, "error due to missing filename")
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
