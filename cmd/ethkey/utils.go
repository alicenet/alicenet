package ethkey

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/alicenet/alicenet/config"
	"github.com/ethereum/go-ethereum/console/prompt"
	"github.com/sirupsen/logrus"
)

// getPassphrase obtains a passphrase given by the user.  It first checks the
// --passfile command line flag and ultimately prompts the user for a
// passphrase.
func getPassphrase(generateRandomPass, confirmation bool, logger *logrus.Entry) string {
	// Look for the --passwordfile flag.
	passphraseFile := config.Configuration.EthKey.PasswordFile
	if passphraseFile != "" {
		content, err := os.ReadFile(passphraseFile)
		if err != nil {
			logger.Fatalf("Failed to read password file '%s': %v",
				passphraseFile, err)
		}
		return strings.TrimRight(string(content), "\r\n")
	}

	// Otherwise prompt the user for the passphrase.
	return getPassPhrase(generateRandomPass, confirmation, logger)
}

// getPassPhrase displays the given text(prompt) to the user and requests some textual
// data to be entered, but one which must not be echoed out into the terminal.
// The method returns the input provided by the user.
func getPassPhrase(generateRandomPass, withConfirmation bool, logger *logrus.Entry) string {
	var err error
	var password string

	if generateRandomPass {
		return getPassword()
	} else {
		password, err = prompt.Stdin.PromptPassword("Password: ")
		if err != nil {
			logger.Fatalf("Failed to read password: %v", err)
		}
		if withConfirmation {
			confirm, err := prompt.Stdin.PromptPassword("Repeat password: ")
			if err != nil {
				logger.Fatalf("Failed to read password confirmation: %v", err)
			}
			if password != confirm {
				logger.Fatalf("Passwords do not match")
			}
		}
	}

	return password
}

// getPassword generates a random ASCII string with at least one digit and one special character.
func getPassword() string {
	rand.Seed(time.Now().UnixNano())
	digits := "0123456789"
	specials := "~=+%^*/()[]{}/!@#$?|"
	all := "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		digits + specials
	length := 8
	buf := make([]byte, length)
	buf[0] = digits[rand.Intn(len(digits))]
	buf[1] = specials[rand.Intn(len(specials))]
	for i := 2; i < length; i++ {
		buf[i] = all[rand.Intn(len(all))]
	}
	rand.Shuffle(len(buf), func(i, j int) {
		buf[i], buf[j] = buf[j], buf[i]
	})
	return string(buf)
}

// generateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
func GenerateRandomString(s int) (string, error) {
	b, err := generateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

// mustPrintJSON prints the JSON encoding of the given object and
// exits the program with an error message when the marshaling fails.
func mustPrintJSON(jsonObject interface{}, logger *logrus.Entry) {
	str, err := json.MarshalIndent(jsonObject, "", "  ")
	if err != nil {
		logger.Fatalf("Failed to marshal JSON object: %v", err)
	}
	fmt.Println(string(str))
}

// getKeyfilePath from args if present
func getKeyfilePath(args []string) (string, bool) {
	if len(args) == 0 || args[0] == "" {
		return "", false
	}

	return args[0], true
}

func ReadYesOrNoAnswer(message string) (bool, error) {
	var result bool

	input, err := ReadInput(message)
	if err != nil {
		return false, err
	}

	input = strings.TrimSuffix(input, "\n")
	if err != nil {
		return false, err
	}

	if input == "" || input == "y" || input == "Y" || input == "yes" || input == "Yes" || input == "YES" {
		result = true
	} else if input != "n" && input != "N" && input != "no" && input != "No" && input != "NO" {
		return false, fmt.Errorf("You entered a wrong answer: %s. Aborting execution", input)
	}

	return result, nil
}

func ReadInput(message string) (string, error) {
	fmt.Print(message)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)

	return input, nil
}
