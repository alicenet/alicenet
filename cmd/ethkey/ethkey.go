package ethkey

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/alicenet/alicenet/config"
	"github.com/alicenet/alicenet/logging"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultKeyfileName = "keyfile.json"
)

type outputGenerate struct {
	Address      string
	AddressEIP55 string
}

// Command is the cobra.Command specifically for generating an ethereum key and address.
var Command = cobra.Command{
	Use:   "generate-ethkey",
	Short: "Generate a new keyfile",
	Long:  "Generate a new keyfile with an ethereum address and private key",
	Run:   generate,
}

func generate(cmd *cobra.Command, args []string) {
	logger := logging.GetLogger("ethkey")

	// Check if keyfile path given and make sure it doesn't already exist.
	keyfilepath := defaultKeyfileName
	if len(args) > 0 && args[0] != "" {
		keyfilepath = args[0]
	}
	if _, err := os.Stat(keyfilepath); err == nil {
		logger.Fatalf("Keyfile already exists at %s.", keyfilepath)
	} else if !os.IsNotExist(err) {
		logger.Fatalf("Error checking if keyfile exists: %v", err)
	}

	var privateKey *ecdsa.PrivateKey
	var err error
	if file := config.Configuration.EthKey.PrivateKey; file != "" {
		// Load private key from file.
		privateKey, err = crypto.LoadECDSA(file)
		if err != nil {
			logger.Fatalf("Can't load private key: %v", err)
		}
	} else {
		// If not loaded, generate random.
		privateKey, err = crypto.GenerateKey()
		if err != nil {
			logger.Fatalf("Failed to generate random private key: %v", err)
		}
	}

	// Create the keyfile object with a random UUID.
	UUID, err := uuid.NewRandom()
	if err != nil {
		logger.Fatalf("Failed to generate random uuid: %v", err)
	}
	key := &keystore.Key{
		Id:         UUID,
		Address:    crypto.PubkeyToAddress(privateKey.PublicKey),
		PrivateKey: privateKey,
	}

	// Encrypt key with passphrase.
	passphrase := getPassphrase(true, logger)
	scryptN, scryptP := keystore.StandardScryptN, keystore.StandardScryptP
	if config.Configuration.EthKey.LightKDF {
		scryptN, scryptP = keystore.LightScryptN, keystore.LightScryptP
	}
	keyjson, err := keystore.EncryptKey(key, passphrase, scryptN, scryptP)
	if err != nil {
		logger.Fatalf("Error encrypting key: %v", err)
	}

	// Store the file to disk.
	if err := os.MkdirAll(filepath.Dir(keyfilepath), 0700); err != nil {
		logger.Fatalf("Could not create directory %s", filepath.Dir(keyfilepath))
	}
	if err := os.WriteFile(keyfilepath, keyjson, 0600); err != nil {
		logger.Fatalf("Failed to write keyfile to %s: %v", keyfilepath, err)
	}

	// Output some information.
	out := outputGenerate{
		Address: key.Address.Hex(),
	}
	if config.Configuration.EthKey.Json {
		mustPrintJSON(out, logger)
	} else {
		fmt.Println("Address:", out.Address)
	}
}

// getPassphrase obtains a passphrase given by the user.  It first checks the
// --passfile command line flag and ultimately prompts the user for a
// passphrase.
func getPassphrase(confirmation bool, logger *logrus.Logger) string {
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
func mustPrintJSON(jsonObject interface{}, logger *logrus.Logger) {
	str, err := json.MarshalIndent(jsonObject, "", "  ")
	if err != nil {
		logger.Fatalf("Failed to marshal JSON object: %v", err)
	}
	fmt.Println(string(str))
}
