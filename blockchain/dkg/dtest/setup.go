package dtest

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

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	"github.com/alicenet/alicenet/blockchain"
	dkgMath "github.com/alicenet/alicenet/blockchain/dkg/math"
	"github.com/alicenet/alicenet/blockchain/etest"
	"github.com/alicenet/alicenet/blockchain/interfaces"
	"github.com/alicenet/alicenet/blockchain/monitor"
	"github.com/alicenet/alicenet/blockchain/objects"
	"github.com/alicenet/alicenet/bridge/bindings"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/crypto/bn256"
	"github.com/alicenet/alicenet/crypto/bn256/cloudflare"
)

type nullWriter struct{}

func (nullWriter) Write(p []byte) (n int, err error) { return len(p), nil }

func InitializeNewDetDkgStateInfo(n int) ([]*objects.DkgState, []*ecdsa.PrivateKey) {
	return InitializeNewDkgStateInfo(n, true)
}

func InitializeNewNonDetDkgStateInfo(n int) ([]*objects.DkgState, []*ecdsa.PrivateKey) {
	return InitializeNewDkgStateInfo(n, false)
}

func InitializeNewDkgStateInfo(n int, deterministicShares bool) ([]*objects.DkgState, []*ecdsa.PrivateKey) {
	// Get private keys for validators
	privKeys := etest.SetupPrivateKeys(n)
	accountsArray := etest.SetupAccounts(privKeys)
	dkgStates := []*objects.DkgState{}
	threshold := crypto.CalcThreshold(n)

	// Make base for secret key
	baseSecretBytes := make([]byte, 32)
	baseSecretBytes[0] = 101
	baseSecretBytes[31] = 101
	baseSecretValue := new(big.Int).SetBytes(baseSecretBytes)

	// Make base for transport key
	baseTransportBytes := make([]byte, 32)
	baseTransportBytes[0] = 1
	baseTransportBytes[1] = 1
	baseTransportValue := new(big.Int).SetBytes(baseTransportBytes)

	// Beginning dkgState initialization
	for k := 0; k < n; k++ {
		bigK := big.NewInt(int64(k))
		// Get base DkgState
		dkgState := objects.NewDkgState(accountsArray[k])
		// Set Index
		dkgState.Index = k + 1
		// Set Number of Validators
		dkgState.NumberOfValidators = n
		dkgState.ValidatorThreshold = threshold

		// Setup TransportKey
		transportPrivateKey := new(big.Int).Add(baseTransportValue, bigK)
		dkgState.TransportPrivateKey = transportPrivateKey
		transportPublicKeyG1 := new(cloudflare.G1).ScalarBaseMult(dkgState.TransportPrivateKey)
		transportPublicKey, err := bn256.G1ToBigIntArray(transportPublicKeyG1)
		if err != nil {
			panic(err)
		}
		dkgState.TransportPublicKey = transportPublicKey

		// Append to state array
		dkgStates = append(dkgStates, dkgState)
	}

	// Generate Participants
	for k := 0; k < n; k++ {
		participantList := GenerateParticipantList(dkgStates)
		for _, p := range participantList {
			dkgStates[k].Participants[p.Address] = p
		}
	}

	// Prepare secret shares
	for k := 0; k < n; k++ {
		bigK := big.NewInt(int64(k))
		// Set SecretValue and PrivateCoefficients
		dkgState := dkgStates[k]
		if deterministicShares {
			// Deterministic shares
			secretValue := new(big.Int).Add(baseSecretValue, bigK)
			privCoefs := GenerateDeterministicPrivateCoefficients(n)
			privCoefs[0].Set(secretValue) // Overwrite constant term
			dkgState.SecretValue = secretValue
			dkgState.PrivateCoefficients = privCoefs
		} else {
			// Random shares
			_, privCoefs, _, err := dkgMath.GenerateShares(dkgState.TransportPrivateKey, dkgState.GetSortedParticipants())
			if err != nil {
				panic(err)
			}
			dkgState.SecretValue = new(big.Int)
			dkgState.SecretValue.Set(privCoefs[0])
			dkgState.PrivateCoefficients = privCoefs
		}
	}

	return dkgStates, privKeys
}

func GenerateParticipantList(dkgStates []*objects.DkgState) objects.ParticipantList {
	n := len(dkgStates)
	participants := make(objects.ParticipantList, int(n))
	for idx := 0; idx < n; idx++ {
		addr := dkgStates[idx].Account.Address
		publicKey := [2]*big.Int{}
		publicKey[0] = new(big.Int)
		publicKey[1] = new(big.Int)
		publicKey[0].Set(dkgStates[idx].TransportPublicKey[0])
		publicKey[1].Set(dkgStates[idx].TransportPublicKey[1])
		participant := &objects.Participant{}
		participant.Address = addr
		participant.PublicKey = publicKey
		participant.Index = dkgStates[idx].Index
		participants[idx] = participant
	}
	return participants
}

func GenerateEncryptedSharesAndCommitments(dkgStates []*objects.DkgState) {
	n := len(dkgStates)
	for k := 0; k < n; k++ {
		dkgState := dkgStates[k]
		publicCoefs := GeneratePublicCoefficients(dkgState.PrivateCoefficients)
		encryptedShares := GenerateEncryptedShares(dkgStates, k)
		// Loop through entire list and save in map
		for ell := 0; ell < n; ell++ {
			dkgStates[ell].Participants[dkgState.Account.Address].Commitments = publicCoefs
			dkgStates[ell].Participants[dkgState.Account.Address].EncryptedShares = encryptedShares
		}
	}
}

func GenerateDeterministicPrivateCoefficients(n int) []*big.Int {
	threshold := crypto.CalcThreshold(n)
	privCoefs := []*big.Int{}
	privCoefs = append(privCoefs, big.NewInt(0))
	for k := 1; k <= threshold; k++ {
		privCoef := big.NewInt(1)
		privCoefs = append(privCoefs, privCoef)
	}
	return privCoefs
}

func GeneratePublicCoefficients(privCoefs []*big.Int) [][2]*big.Int {
	publicCoefsG1 := cloudflare.GeneratePublicCoefs(privCoefs)
	publicCoefs := [][2]*big.Int{}
	for k := 0; k < len(publicCoefsG1); k++ {
		coefG1 := publicCoefsG1[k]
		coef, err := bn256.G1ToBigIntArray(coefG1)
		if err != nil {
			panic(err)
		}
		publicCoefs = append(publicCoefs, coef)
	}
	return publicCoefs
}

func GenerateEncryptedShares(dkgStates []*objects.DkgState, idx int) []*big.Int {
	dkgState := dkgStates[idx]
	// Get array of public keys and convert to cloudflare.G1
	publicKeysBig := [][2]*big.Int{}
	for k := 0; k < len(dkgStates); k++ {
		publicKeysBig = append(publicKeysBig, dkgStates[k].TransportPublicKey)
	}
	publicKeysG1, err := bn256.BigIntArraySliceToG1(publicKeysBig)
	if err != nil {
		panic(err)
	}

	// Get public key for caller
	publicKeyBig := dkgState.TransportPublicKey
	publicKey, err := bn256.BigIntArrayToG1(publicKeyBig)
	if err != nil {
		panic(err)
	}
	privCoefs := dkgState.PrivateCoefficients
	secretShares, err := cloudflare.GenerateSecretShares(publicKey, privCoefs, publicKeysG1)
	if err != nil {
		panic(err)
	}
	encryptedShares, err := cloudflare.GenerateEncryptedShares(secretShares, dkgState.TransportPrivateKey, publicKeysG1)
	if err != nil {
		panic(err)
	}
	return encryptedShares
}

func GenerateKeyShares(dkgStates []*objects.DkgState) {
	n := len(dkgStates)
	for k := 0; k < n; k++ {
		dkgState := dkgStates[k]
		g1KeyShare, g1Proof, g2KeyShare, err := dkgMath.GenerateKeyShare(dkgState.SecretValue)
		if err != nil {
			panic(err)
		}
		addr := dkgState.Account.Address
		// Loop through entire list and save in map
		for ell := 0; ell < n; ell++ {
			dkgStates[ell].Participants[addr].KeyShareG1s = g1KeyShare
			dkgStates[ell].Participants[addr].KeyShareG1CorrectnessProofs = g1Proof
			dkgStates[ell].Participants[addr].KeyShareG2s = g2KeyShare
		}
	}
}

// GenerateMasterPublicKey computes the mpk for the protocol.
// This computes this by using all of the secret values from dkgStates.
func GenerateMasterPublicKey(dkgStates []*objects.DkgState) []*objects.DkgState {
	n := len(dkgStates)
	msk := new(big.Int)
	for k := 0; k < n; k++ {
		msk.Add(msk, dkgStates[k].SecretValue)
	}
	msk.Mod(msk, cloudflare.Order)
	for k := 0; k < n; k++ {
		mpkG2 := new(cloudflare.G2).ScalarBaseMult(msk)
		mpk, err := bn256.G2ToBigIntArray(mpkG2)
		if err != nil {
			panic(err)
		}
		dkgStates[k].MasterPublicKey = mpk
	}
	return dkgStates
}

func GenerateGPKJ(dkgStates []*objects.DkgState) {
	n := len(dkgStates)
	for k := 0; k < n; k++ {
		dkgState := dkgStates[k]

		encryptedShares := make([][]*big.Int, n)
		for idx, participant := range dkgState.GetSortedParticipants() {
			p, present := dkgState.Participants[participant.Address]
			if present && idx >= 0 && idx < n {
				encryptedShares[idx] = p.EncryptedShares
			} else {
				panic("Encrypted share state broken")
			}
		}

		groupPrivateKey, groupPublicKey, err := dkgMath.GenerateGroupKeys(dkgState.TransportPrivateKey, dkgState.PrivateCoefficients,
			encryptedShares, dkgState.Index, dkgState.GetSortedParticipants())
		if err != nil {
			panic("Could not generate group keys")
		}

		dkgState.GroupPrivateKey = groupPrivateKey

		// Loop through entire list and save in map
		for ell := 0; ell < n; ell++ {
			dkgStates[ell].Participants[dkgState.Account.Address].GPKj = groupPublicKey
		}
	}
}

func GetAliceNetRootPath() []string {

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
	randomPrivateKeys := etest.SetupPrivateKeys(n - 1)
	privateKeys = append(privateKeys, randomPrivateKeys...)
	accounts := etest.SetupAccounts(privateKeys)

	return privateKeys, accounts
}

func ReadFromFileOnRoot(filePath string, configVar string) (string, error) {
	rootPath := GetAliceNetRootPath()
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
	rootPath := GetAliceNetRootPath()

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

func ConnectSimulatorEndpoint(t *testing.T, privateKeys []*ecdsa.PrivateKey, blockInterval time.Duration) interfaces.Ethereum {
	eth, err := blockchain.NewEthereumSimulator(
		privateKeys,
		6,
		10*time.Second,
		30*time.Second,
		0,
		big.NewInt(math.MaxInt64),
		50,
		math.MaxInt64,
		5*time.Second,
		30*time.Second)

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
		rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
		assert.Nil(t, err)
		assert.NotNil(t, rcpt)
	}

	return eth
}

func StartHardHatNode(eth *blockchain.EthereumDetails) error {

	rootPath := GetAliceNetRootPath()
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

		//err = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		//if err != nil {
		//	return err
		//}

		//err = cmd.Process.Kill()
		//if err != nil {
		//	return err
		//}

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
		cmd.Stdout = nullWriter{}
		cmd.Stderr = nullWriter{}
	}
}

func InitializeValidatorFiles(n int) error {

	rootPath := GetAliceNetRootPath()
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

func StartDeployScripts(eth *blockchain.EthereumDetails, ctx context.Context) error {

	rootPath := GetAliceNetRootPath()
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
	err = eth.SetContractFactory(ctx, addr)
	if err != nil {
		return err
	}

	return nil
}

func WaitForHardHatNode(ctx context.Context) error {
	c := http.Client{}
	msg := &blockchain.JsonrpcMessage{
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
		log.Fatal(err)
	}

	for {
		reader := bytes.NewReader(buff.Bytes())

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
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}

	return err
}

func RegisterValidators(eth *blockchain.EthereumDetails, validatorAddresses []string) error {

	rootPath := GetAliceNetRootPath()
	scriptPath := append(rootPath, "scripts")
	scriptPath = append(scriptPath, "main.sh")
	scriptPathJoined := filepath.Join(scriptPath...)
	fmt.Println("scriptPathJoined: ", scriptPathJoined)

	args := []string{
		scriptPathJoined,
		"register_test",
		eth.Contracts().ContractFactoryAddress().String(),
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

func GetETHDKGRegistrationOpened(logs []*types.Log, eth interfaces.Ethereum) (*bindings.ETHDKGRegistrationOpened, error) {
	eventMap := monitor.GetETHDKGEvents()
	eventInfo, ok := eventMap["RegistrationOpened"]
	if !ok {
		return nil, fmt.Errorf("event not found: %v", eventInfo.Name)
	}

	var event *bindings.ETHDKGRegistrationOpened
	var err error
	for _, log := range logs {
		for _, topic := range log.Topics {
			if topic.String() == eventInfo.ID.String() {
				event, err = eth.Contracts().Ethdkg().ParseRegistrationOpened(*log)
				if err != nil {
					continue
				}

				return event, nil
			}
		}
	}
	return nil, fmt.Errorf("event not found")
}
