package tests

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"

	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/alicenet/alicenet/layer1/transaction"
	"github.com/alicenet/alicenet/utils"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
)

// SetupPrivateKeys computes deterministic private keys for testing
func SetupPrivateKeys(n int) []*ecdsa.PrivateKey {
	if (n < 1) || (n >= 256) {
		panic("invalid number for accounts")
	}
	secp256k1N, _ := new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	baseBytes := make([]byte, 32)
	baseBytes[0] = 255
	baseBytes[31] = 255
	privKeyArray := []*ecdsa.PrivateKey{}
	for k := 0; k < n; k++ {
		privKeyBytes := utils.CopySlice(baseBytes)
		privKeyBytes[1] = uint8(k)
		privKeyBig := new(big.Int).SetBytes(privKeyBytes)
		privKeyBig.Mod(privKeyBig, secp256k1N)
		privKeyBytes = privKeyBig.Bytes()
		privKey, err := crypto.ToECDSA(privKeyBytes)
		if err != nil {
			panic(err)
		}
		privKeyArray = append(privKeyArray, privKey)
	}
	return privKeyArray
}

// CreateAccounts creates the privateKeys and accounts for a given number of
// accounts. The first created account will be always the hardhat admin account.
func CreateAccounts(unitTestDirectory string, numAccounts int) (string, string) {
	if numAccounts < 1 {
		panic(fmt.Errorf("The number of accounts must be greater than 1, given number %v", numAccounts))
	}
	keyStorePath := filepath.Join(unitTestDirectory, "keys")
	passCodePath := filepath.Join(unitTestDirectory, "passcodes.txt")
	keystore := keystore.NewKeyStore(keyStorePath, keystore.StandardScryptN, keystore.StandardScryptP)
	privateKeys := SetupPrivateKeys(numAccounts)
	passCodesFile, err := os.OpenFile(passCodePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Errorf("failed to open/create passCode file: %v", err))
	}
	defer passCodesFile.Close()
	for _, privateKey := range privateKeys {
		account, err := keystore.ImportECDSA(privateKey, "abc123")
		if err != nil {
			panic(fmt.Errorf("failed to import private key: %v", err))
		}
		if _, err := passCodesFile.WriteString(fmt.Sprintf("%v=abc123\n", account.Address.Hex())); err != nil {
			panic(fmt.Errorf("failed to save passCode in the passCode file: %v", err))
		}
	}
	return keyStorePath, passCodePath
}

func InitializePrivateKeys(n int) []*ecdsa.PrivateKey {
	_, pKey := GetAdminAccount()
	privateKeys := []*ecdsa.PrivateKey{pKey}
	privateKeys = append(privateKeys, SetupPrivateKeys(n-1)...)

	return privateKeys
}

// GetAdminAccount gets the admin account for the hardhat node. If that admin account is changed in the hardhat configs change this.
func GetAdminAccount() (common.Address, *ecdsa.PrivateKey) {
	privateKey, err := crypto.ToECDSA(common.Hex2Bytes(TestAdminPrivateKey))
	if err != nil {
		panic(fmt.Errorf("failed to get test admin privatekey: %v", err))
	}
	return crypto.PubkeyToAddress(privateKey.PublicKey), privateKey
}

func fundAccounts(eth layer1.Client, watcher transaction.Watcher, logger *logrus.Entry) error {
	knownAccounts := eth.GetKnownAccounts()
	var receiptResponses []transaction.ReceiptResponse
	// transferring 100 ether
	amount := new(big.Int).Mul(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil), big.NewInt(100))
	// fund known accounts
	for _, account := range knownAccounts {
		txn, err := ethereum.TransferEther(
			eth,
			logger.WithField("account", account),
			eth.GetDefaultAccount().Address,
			account.Address,
			amount,
		)
		if err != nil {
			return fmt.Errorf("failed to fund account: %v", account.Address.Hex())
		}
		rcptResponse, err := watcher.Subscribe(context.Background(), txn, nil)
		if err != nil {
			return fmt.Errorf("failed to subscribe txn for account: %v", account.Address.Hex())
		}
		receiptResponses = append(receiptResponses, rcptResponse)
	}
	MineBlocks(eth.GetEndpoint(), eth.GetFinalityDelay()+1)
	// all receipts should be ready
	for index, rcptResponse := range receiptResponses {
		if !rcptResponse.IsReady() {
			return fmt.Errorf("failed to get receipt for account: %v", knownAccounts[index].Address.Hex())
		}
	}
	return nil
}

type ClientFixture struct {
	Client         layer1.Client
	Watcher        transaction.Watcher
	FactoryAddress string
	TempDir        string
	KeyStorePath   string
	PassCodePath   string
	Logger         *logrus.Entry
}

func NewClientFixture(hardhat *Hardhat, finalityDelay uint64, numAccounts int, logger *logrus.Entry, unlockAllAccounts bool, deployContracts bool, registerValidators bool) *ClientFixture {
	tempDir, err := os.MkdirTemp("", "unittestdir")
	if err != nil {
		panic(fmt.Errorf("failed to create tmp dir: %v", err))
	}
	keyStorePath, passCodePath := CreateAccounts(tempDir, numAccounts)
	defaultAccount, _ := GetAdminAccount()
	eth, err := ethereum.NewClient(
		hardhat.url,
		keyStorePath,
		passCodePath,
		defaultAccount.Hex(),
		unlockAllAccounts,
		finalityDelay,
		math.MaxInt64,
		0,
	)
	if err != nil {
		panic(fmt.Errorf("failed to create ethereum client: %v", err))
	}

	ResetHardhatConfigs(hardhat.url)

	watcher := transaction.WatcherFromNetwork(eth, nil, false)

	err = fundAccounts(eth, watcher, logger)
	if err != nil {
		panic(fmt.Errorf("failed to fund validators: %v", err))
	}

	factoryAddress := ""
	if deployContracts {
		baseFilesDir := filepath.Join(GetProjectRootPath(), "scripts", "base-files")
		factoryAddress, err = hardhat.DeployFactoryAndContracts(baseFilesDir)
		if err != nil {
			panic(fmt.Errorf("failed to deploy factory: %v", err))
		}

		var validatorsAddresses []string
		for _, account := range eth.GetKnownAccounts() {
			validatorsAddresses = append(validatorsAddresses, account.Address.Hex())
		}
		if registerValidators {
			hardhat.RegisterValidators(factoryAddress, validatorsAddresses)
		}
	}

	return &ClientFixture{
		Client:         eth,
		Watcher:        watcher,
		TempDir:        tempDir,
		FactoryAddress: factoryAddress,
		PassCodePath:   passCodePath,
		KeyStorePath:   keyStorePath,
	}
}

func (c *ClientFixture) Close() {
	c.Watcher.Close()
	ResetHardhatConfigs(c.Client.GetEndpoint())
	c.Client.Close()
	os.RemoveAll(c.TempDir)
}
