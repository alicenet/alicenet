package ethkey

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alicenet/alicenet/config"
	"github.com/alicenet/alicenet/logging"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	defaultKeyfileName = "keyfile.json"
)

type outputGenerate struct {
	Address      string
	AddressEIP55 string
}

// Generate is the cobra.Command specifically for generating an ethereum key and address.
var Generate = cobra.Command{
	Use:   "ethkey-generate",
	Short: "Generate a new keyfile",
	Long:  "Generate a new keyfile with an ethereum address and private key",
	Run:   generate,
}

func generate(cmd *cobra.Command, args []string) {
	logger := logging.GetLogger("ethkey").WithField("method", "generate")

	// Check if keyfile path given and make sure it doesn't already exist.
	keyFilePath := defaultKeyfileName
	if kfp, ok := getKeyfilePath(args); ok {
		keyFilePath = kfp
	}
	if _, err := os.Stat(keyFilePath); err == nil {
		logger.Fatalf("Keyfile already exists at %s.", keyFilePath)
	} else if !os.IsNotExist(err) {
		logger.Fatalf("Error checking if keyfile exists: %v", err)
	}

	keyjson, key, _, err := GenerateKeyFile(false, logger)
	if err != nil {
		logger.Fatalf(err.Error())
	}

	// Store the file to disk.
	if err := os.MkdirAll(filepath.Dir(keyFilePath), 0700); err != nil {
		logger.Fatalf("Could not create directory %s", filepath.Dir(keyFilePath))
	}
	if err := os.WriteFile(keyFilePath, keyjson, 0600); err != nil {
		logger.Fatalf("Failed to write keyfile to %s: %v", keyFilePath, err)
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

func GenerateKeyFile(generateRandomPass bool, logger *logrus.Entry) ([]byte, *keystore.Key, string, error) {
	var privateKey *ecdsa.PrivateKey
	var err error
	if file := config.Configuration.EthKey.PrivateKey; file != "" {
		// Load private key from file.
		privateKey, err = crypto.LoadECDSA(file)
		if err != nil {
			return nil, nil, "", fmt.Errorf("Can't load private key: %v", err)
		}
	} else {
		// If not loaded, generate random.
		privateKey, err = crypto.GenerateKey()
		if err != nil {
			return nil, nil, "", fmt.Errorf("Failed to generate random private key: %v", err)
		}
	}

	// Create the keyfile object with a random UUID.
	UUID, err := uuid.NewRandom()
	if err != nil {
		return nil, nil, "", fmt.Errorf("Failed to generate random uuid: %v", err)
	}
	key := &keystore.Key{
		Id:         UUID,
		Address:    crypto.PubkeyToAddress(privateKey.PublicKey),
		PrivateKey: privateKey,
	}

	// Encrypt key with passphrase.
	passphrase := getPassphrase(generateRandomPass, true, logger)
	scryptN, scryptP := keystore.StandardScryptN, keystore.StandardScryptP
	if config.Configuration.EthKey.LightKDF {
		scryptN, scryptP = keystore.LightScryptN, keystore.LightScryptP
	}
	keyjson, err := keystore.EncryptKey(key, passphrase, scryptN, scryptP)
	if err != nil {
		return nil, nil, "", fmt.Errorf("Error encrypting key: %v", err)
	}

	return keyjson, key, passphrase, nil
}
