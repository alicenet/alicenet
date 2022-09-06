package ethkey

import (
	"encoding/json"
	"fmt"
	"github.com/alicenet/alicenet/config"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

// getPassphrase obtains a passphrase given by the user.  It first checks the
// --passfile command line flag and ultimately prompts the user for a
// passphrase.
func getPassphrase(confirmation bool, logger *logrus.Entry) string {
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
	return utils.GetPassPhrase("", confirmation)
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
