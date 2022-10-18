package ethkey

import (
	"fmt"
	"github.com/alicenet/alicenet/config"
	"github.com/alicenet/alicenet/logging"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

// ChangePassword is the cobra.Command specifically for changing the password for a keyfile.
var ChangePassword = cobra.Command{
	Use:   "ethkey-changepassword",
	Short: "Change the password of a keyfile",
	Long:  "Change the password of a keyfile. use the --newpasswordfile to point to the new password file.",
	Run:   changePassword,
}

func changePassword(cmd *cobra.Command, args []string) {
	logger := logging.GetLogger("ethkey").WithField("method", "changepassword")

	keyFilePath, ok := getKeyfilePath(args)
	if !ok {
		logger.Fatalf("The keyfile wasn't specified")
	}

	// Read key from file.
	keyjson, err := os.ReadFile(keyFilePath)
	if err != nil {
		logger.Fatalf("Failed to read the keyfile at '%s': %v", keyFilePath, err)
	}

	// Decrypt key with passphrase.
	passphrase := getPassphrase(false, logger)
	key, err := keystore.DecryptKey(keyjson, passphrase)
	if err != nil {
		logger.Fatalf("Error decrypting key: %v", err)
	}

	// Get a new passphrase.
	fmt.Println("Please provide a new password")
	var newPhrase string
	if newPassFile := config.Configuration.EthKey.NewPasswordFile; newPassFile != "" {
		content, err := os.ReadFile(newPassFile)
		if err != nil {
			logger.Fatalf("Failed to read new password file '%s': %v", newPassFile, err)
		}
		newPhrase = strings.TrimRight(string(content), "\r\n")
	} else {
		newPhrase = utils.GetPassPhrase("", true)
	}

	// Encrypt the key with the new passphrase.
	newJson, err := keystore.EncryptKey(key, newPhrase, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		logger.Fatalf("Error encrypting with new password: %v", err)
	}

	// Then write the new keyfile in place of the old one.
	if err := os.WriteFile(keyFilePath, newJson, 0600); err != nil {
		logger.Fatalf("Error writing new keyfile to disk: %v", err)
	}
}
