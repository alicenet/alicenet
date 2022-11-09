package ethkey

import (
	"encoding/hex"
	"fmt"
	"github.com/alicenet/alicenet/config"
	"github.com/alicenet/alicenet/logging"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
	"os"
)

// Inspect is the cobra.Command specifically decrypting the keyfile address, public and private keys.
var Inspect = cobra.Command{
	Use:   "ethkey-inspect",
	Short: "Print various information about the keyfile",
	Long:  "Print various information about the keyfile. Private key information can be printed by using the --private flag; make sure to use this feature with great caution!",
	Run:   inspect,
}

type outputInspect struct {
	Address    string
	PublicKey  string
	PrivateKey string
}

func inspect(cmd *cobra.Command, args []string) {
	logger := logging.GetLogger("ethkey").WithField("method", "inspect")

	keyFilePath, ok := getKeyfilePath(args)
	if !ok {
		logger.Fatalf("The keyfile wasn't specified")
	}

	// Read key from file.
	keyJSON, err := os.ReadFile(keyFilePath)
	if err != nil {
		logger.Fatalf("Failed to read the keyfile at '%s': %v", keyFilePath, err)
	}

	// Decrypt key with passphrase.
	passphrase := getPassphrase(false, false, logger)
	key, err := keystore.DecryptKey(keyJSON, passphrase)
	if err != nil {
		logger.Fatalf("Error decrypting key: %v", err)
	}

	// Output all relevant information we can retrieve.
	showPrivate := config.Configuration.EthKey.Private
	out := outputInspect{
		Address: key.Address.Hex(),
		PublicKey: hex.EncodeToString(
			crypto.FromECDSAPub(&key.PrivateKey.PublicKey)),
	}
	if showPrivate {
		out.PrivateKey = hex.EncodeToString(crypto.FromECDSA(key.PrivateKey))
	}

	if config.Configuration.EthKey.Json {
		mustPrintJSON(out, logger)
	} else {
		fmt.Println("Address:       ", out.Address)
		fmt.Println("Public key:    ", out.PublicKey)
		if showPrivate {
			fmt.Println("Private key:   ", out.PrivateKey)
		}
	}
}
