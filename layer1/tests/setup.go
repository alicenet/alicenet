package tests

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/logging"
	"github.com/sirupsen/logrus"

	"github.com/alicenet/alicenet/blockchain/ethereum"
	"github.com/alicenet/alicenet/layer1/transaction"

	"github.com/alicenet/alicenet/utils"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
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

// SetAccounts derives the associated addresses from private keys
func SetupAccounts(privKeys []*ecdsa.PrivateKey) []accounts.Account {
	accountsArray := []accounts.Account{}
	for _, pk := range privKeys {
		commonAddr := crypto.PubkeyToAddress(pk.PublicKey)
		accountValue := accounts.Account{Address: commonAddr}
		accountsArray = append(accountsArray, accountValue)
	}
	return accountsArray
}

func GetMadnetRootPath() []string {

	rootPath := []string{string(os.PathSeparator)}

	cmd := exec.Command("go", "list", "-m", "-f", "'{{.Dir}}'", "github.com/alicenet/alicenet")
	stdout, err := cmd.Output()
	if err != nil {
		log.Printf("Error getting project root path: %v", err)
		return rootPath
	}

	path := string(stdout)
	path = strings.ReplaceAll(path, "'", "")
	path = strings.ReplaceAll(path, "\n", "")

	pathNodes := strings.Split(path, string(os.PathSeparator))
	for _, pathNode := range pathNodes {
		rootPath = append(rootPath, pathNode)
	}

	return rootPath
}

func InitializePrivateKeysAndAccounts(n int) ([]*ecdsa.PrivateKey, []accounts.Account) {
	_, pKey, err := GetOwnerAccount()
	if err != nil {
		panic(err)
	}

	//t.Logf("owner: %v, pvKey: %v", account.Address.String(), key.PrivateKey)
	privateKeys := []*ecdsa.PrivateKey{pKey}
	randomPrivateKeys := SetupPrivateKeys(n - 1)
	privateKeys = append(privateKeys, randomPrivateKeys...)
	accounts := SetupAccounts(privateKeys)

	return privateKeys, accounts
}

func ReadFromFileOnRoot(filePath string, configVar string) (string, error) {
	rootPath := GetMadnetRootPath()
	rootPath = append(rootPath, filePath)
	fileFullPath := filepath.Join(rootPath...)

	f, err := os.Open(fileFullPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Splits on newlines by default.
	scanner := bufio.NewScanner(f)
	var defaultAccount string

	// https://golang.org/pkg/bufio/#Scanner.Scan
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), configVar) {
			defaultAccount = scanner.Text()
			break
		}
	}

	splits := strings.Split(defaultAccount, "=")
	return strings.Trim(splits[1], " \""), nil
}

func GetOwnerAccount() (*common.Address, *ecdsa.PrivateKey, error) {
	rootPath := GetMadnetRootPath()

	// open config file owner.toml
	acctAddress, err := ReadFromFileOnRoot("scripts/base-files/owner.toml", "defaultAccount")
	if err != nil {
		return nil, nil, err
	}
	acctAddressLowerCase := strings.ToLower(acctAddress)

	// open password file
	passwordPath := append(rootPath, "scripts")
	passwordPath = append(passwordPath, "base-files")
	passwordPath = append(passwordPath, "passwordFile")
	passwordFullPath := filepath.Join(passwordPath...)

	fileContent, err := ioutil.ReadFile(passwordFullPath)
	if err != nil {
		//log.Errorf("error opening passsword file: %v", err)
		panic(err)
	}

	// Convert []byte to string
	password := string(fileContent)

	// open wallet json file
	walletPath := append(rootPath, "scripts")
	walletPath = append(walletPath, "base-files")
	walletPath = append(walletPath, acctAddressLowerCase)
	walletFullPath := filepath.Join(walletPath...)

	jsonBytes, err := ioutil.ReadFile(walletFullPath)
	if err != nil {
		panic(err)
	}

	key, err := keystore.DecryptKey(jsonBytes, password)
	if err != nil {
		panic(err)
	}

	return &key.Address, key.PrivateKey, nil
}

func ConnectSimulatorEndpoint(t *testing.T, privateKeys []*ecdsa.PrivateKey, blockInterval time.Duration) layer1.Client {
	eth, err := ethereum.NewSimulator(
		privateKeys,
		0,
		big.NewInt(math.MaxInt64),
		math.MaxInt64)

	assert.Nil(t, err, "Failed to build Ethereum endpoint...")
	// assert.True(t, eth.IsEthereumAccessible(), "Web3 endpoint is not available.")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Unlock the default account and use it to deploy contracts
	deployAccount := eth.GetDefaultAccount()
	err = eth.UnlockAccount(deployAccount)
	assert.Nil(t, err, "Failed to unlock default account")

	t.Logf("deploy account: %v", deployAccount.Address.String())

	err = StartHardHatNode(eth)
	if err != nil {
		eth.Close()
		t.Fatalf("error starting hardhat node: %v", err)
	}

	t.Logf("waiting on hardhat node to start...")

	err = WaitForHardHatNode(ctx)
	if err != nil {
		eth.Close()
		t.Fatalf("error: %v", err)
	}

	t.Logf("deploying contracts..")

	err = StartDeployScripts(eth, ctx)
	if err != nil {
		eth.Close()
		t.Fatalf("error deploying: %v", err)
	}

	validatorAddresses := make([]string, 0)
	for _, acct := range eth.GetKnownAccounts() {
		validatorAddresses = append(validatorAddresses, acct.Address.String())
	}

	err = RegisterValidators(eth, validatorAddresses)
	assert.Nil(t, err)

	// unlock accounts
	for _, account := range eth.GetKnownAccounts() {
		err := eth.UnlockAccount(account)
		assert.Nil(t, err)
	}

	// fund accounts
	for _, account := range eth.GetKnownAccounts()[1:] {
		txn, err := eth.TransferEther(deployAccount.Address, account.Address, big.NewInt(100000000000000000))
		assert.Nil(t, err)
		assert.NotNil(t, txn)
		if txn == nil {
			// this shouldn't be needed, but is
			eth.Close()
			t.Fatal("could not transfer ether")
		}
		watcher := transaction.NewWatcher(eth.GetClient(), transaction.NewKnownSelectors(), eth.GetFinalityDelay())
		watcher.StartLoop()

		rcpt, err := watcher.SubscribeAndWait(ctx, txn)
		assert.Nil(t, err)
		assert.NotNil(t, rcpt)
	}

	return eth
}

func Setup(finalityDelay uint64, numAccounts int, registryAddress common.Address) (layer1.Client, *logrus.Logger, error) {
	logger := logging.GetLogger("test")
	logger.SetLevel(logrus.TraceLevel)
	ecdsaPrivateKeys, _ := InitializePrivateKeysAndAccounts(numAccounts)
	eth, err := ethereum.NewSimulator(
		ecdsaPrivateKeys,
		6,
		10*time.Second,
		30*time.Second,
		0,
		big.NewInt(math.MaxInt64),
		50,
		math.MaxInt64)
	if err != nil {
		return nil, logger, err
	}

	eth.SetFinalityDelay(finalityDelay)
	knownSelectors := transaction.NewKnownSelectors()
	transaction := transaction.NewWatcher(eth.GetClient(), knownSelectors, 5)
	transaction.SetNumOfConfirmationBlocks(finalityDelay)

	//todo: redeploy and get the registryAddress here
	err = ethereum.GetContracts().LookupContracts(context.Background(), registryAddress)
	if err != nil {
		return nil, logger, err
	}
	return eth, logger, nil
}
func StartHardHatNode(eth *ethereum.Client) error {

	rootPath := GetMadnetRootPath()
	scriptPath := append(rootPath, "scripts")
	scriptPath = append(scriptPath, "main.sh")
	scriptPathJoined := filepath.Join(scriptPath...)
	fmt.Println("scriptPathJoined2: ", scriptPathJoined)

	cmd := exec.Cmd{
		Path: scriptPathJoined,
		Args: []string{scriptPathJoined, "hardhat_node"},
		Dir:  filepath.Join(rootPath...),
	}

	setCommandStdOut(&cmd)
	err := cmd.Start()

	// if there is an error with our execution
	// handle it here
	if err != nil {
		return fmt.Errorf("could not run hardhat node: %s", err)
	}

	eth.SetClose(func() error {
		fmt.Printf("closing hardhat node %v..\n", cmd.Process.Pid)
		err := cmd.Process.Signal(syscall.SIGTERM)
		if err != nil {
			return err
		}

		_, err = cmd.Process.Wait()
		if err != nil {
			return err
		}

		fmt.Printf("hardhat node closed\n")
		return nil
	})

	return nil
}

// setCommandStdOut If ENABLE_SCRIPT_LOG env variable is set as 'true' the command will show scripts logs
func setCommandStdOut(cmd *exec.Cmd) {

	flagValue, found := os.LookupEnv("ENABLE_SCRIPT_LOG")
	enabled, err := strconv.ParseBool(flagValue)

	if err == nil && found && enabled {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	}
}

func InitializeValidatorFiles(n int) error {

	rootPath := GetMadnetRootPath()
	scriptPath := append(rootPath, "scripts")
	scriptPath = append(scriptPath, "main.sh")
	scriptPathJoined := filepath.Join(scriptPath...)
	fmt.Println("scriptPathJoined2: ", scriptPathJoined)

	cmd := exec.Cmd{
		Path: scriptPathJoined,
		Args: []string{scriptPathJoined, "init", strconv.Itoa(n)},
		Dir:  filepath.Join(rootPath...),
	}

	setCommandStdOut(&cmd)
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("could not generate validator files: %s", err)
	}

	return nil
}

func StartDeployScripts(eth *ethereum.Client, ctx context.Context) error {

	rootPath := GetMadnetRootPath()
	scriptPath := append(rootPath, "scripts")
	scriptPath = append(scriptPath, "main.sh")
	scriptPathJoined := filepath.Join(scriptPath...)
	fmt.Println("scriptPathJoined: ", scriptPathJoined)

	err := os.Setenv("SKIP_REGISTRATION", "1")
	if err != nil {
		return err
	}

	cmd := exec.Cmd{
		Path: scriptPathJoined,
		Args: []string{scriptPathJoined, "deploy"},
		Dir:  filepath.Join(rootPath...),
	}

	setCommandStdOut(&cmd)
	err = cmd.Run()

	// if there is an error with our execution
	// handle it here
	if err != nil {
		log.Printf("Could not execute deploy script: %s", err)
		return err
	}

	// inits contracts
	factory, err := ReadFromFileOnRoot("scripts/generated/factoryState", "defaultFactoryAddress")
	if err != nil {
		return err
	}

	addr := common.Address{}
	copy(addr[:], common.FromHex(factory))
	ethereum.GetContracts().Initialize(ctx, addr)

	return nil
}

func WaitForHardHatNode(ctx context.Context) error {
	c := http.Client{}
	msg := &ethereum.JsonRPCMessage{
		Version: "2.0",
		ID:      []byte("1"),
		Method:  "eth_chainId",
		Params:  make([]byte, 0),
	}
	var err error
	if msg.Params, err = json.Marshal(make([]string, 0)); err != nil {
		panic(err)
	}

	var buff bytes.Buffer
	err = json.NewEncoder(&buff).Encode(msg)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(buff.Bytes())

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
			resp, err := c.Post(
				"http://127.0.0.1:8545",
				"application/json",
				reader,
			)
			if err != nil {
				continue
			}
			_, err = io.ReadAll(resp.Body)
			if err == nil {
				return nil
			}
		}

	}
}

func RegisterValidators(eth *ethereum.Client, validatorAddresses []string) error {

	rootPath := GetMadnetRootPath()
	scriptPath := append(rootPath, "scripts")
	scriptPath = append(scriptPath, "main.sh")
	scriptPathJoined := filepath.Join(scriptPath...)
	fmt.Println("scriptPathJoined: ", scriptPathJoined)

	args := []string{
		scriptPathJoined,
		"register_test",
		ethereum.GetContracts().ContractFactoryAddress().String(),
	}
	args = append(args, validatorAddresses...)

	cmd := exec.Cmd{
		Path: scriptPathJoined,
		Args: args,
		Dir:  filepath.Join(rootPath...),
	}

	setCommandStdOut(&cmd)
	err := cmd.Run()

	// if there is an error with our execution
	// handle it here
	if err != nil {
		return fmt.Errorf("could not execute deploy script: %s", err)
	}

	return nil
}

// send a command to the hardhat server via an RPC call
func SendHardhatCommand(command string, params ...interface{}) error {

	commandJson := &ethereum.JsonRPCMessage{
		Version: "2.0",
		ID:      []byte("1"),
		Method:  command,
		Params:  make([]byte, 0),
	}

	paramsJson, err := json.Marshal(params)
	if err != nil {
		return err
	}

	commandJson.Params = paramsJson

	c := http.Client{}
	var buff bytes.Buffer
	err = json.NewEncoder(&buff).Encode(commandJson)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(buff.Bytes())

	resp, err := c.Post(
		"http://127.0.0.1:8545",
		"application/json",
		reader,
	)

	if err != nil {
		return err
	}

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return nil
}

// mine a certain number of hardhat blocks
func MineBlocks(t *testing.T, eth layer1.Client, blocksToMine uint64) {
	var blocksToMineString = "0x" + strconv.FormatUint(blocksToMine, 16)
	log.Printf("hardhat_mine %v blocks ", blocksToMine)
	err := SendHardhatCommand("hardhat_mine", blocksToMineString)
	if err != nil {
		panic(err)
	}
}

// advance to a certain block number
func AdvanceTo(t *testing.T, eth layer1.Client, target uint64) {
	currentBlock, err := eth.GetCurrentHeight(context.Background())
	if err != nil {
		panic(err)
	}
	if target < currentBlock {
		return
	}
	blocksToMine := target - currentBlock
	var blocksToMineString = "0x" + strconv.FormatUint(blocksToMine, 16)

	log.Printf("hardhat_mine %v blocks to target height %v", blocksToMine, target)

	err = SendHardhatCommand("hardhat_mine", blocksToMineString)
	if err != nil {
		panic(err)
	}
}

// The the Base fee for the next hardhat block. Can be used to make tx stale.
func SetNextBlockBaseFee(t *testing.T, eth layer1.Client, target uint64) {
	log.Printf("Setting hardhat_setNextBlockBaseFeePerGas to %v", target)

	err := SendHardhatCommand("hardhat_setNextBlockBaseFeePerGas", "0x"+strconv.FormatUint(target, 16))
	if err != nil {
		panic(err)
	}
}

// Enable/disable hardhat autoMine
func SetAutoMine(t *testing.T, eth layer1.Client, autoMine bool) {
	log.Printf("Setting Automine to %v", autoMine)

	err := SendHardhatCommand("evm_setAutomine", autoMine)
	if err != nil {
		panic(err)
	}
}

// Set the interval between hardhat blocks. In case interval is 0, we enter in
// manual mode and blocks can only be mined explicitly by calling `MineBlocks`.
// This function disables autoMine.
func SetBlockInterval(t *testing.T, eth layer1.Client, intervalInMilliSeconds uint64) {
	SetAutoMine(t, eth, false)
	log.Printf("Setting block interval to %v seconds", intervalInMilliSeconds)
	err := SendHardhatCommand("evm_setIntervalMining", intervalInMilliSeconds)
	if err != nil {
		panic(err)
	}
}
