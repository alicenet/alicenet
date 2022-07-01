package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/ethereum"
	"github.com/alicenet/alicenet/logging"
)

// Test admin private key. DONT USE THIS ON MAINNET!!!!
const (
	TestAdminPrivateKey          string = "6aea45ee1273170fb525da34015e4f20ba39fe792f486ba74020bcacc9badfc1"
	SmartContractsRelativeFolder string = "bridge"
)

func sanitizePathFromOutput(output []byte) string {
	rootPath := []string{string(os.PathSeparator)}
	path := string(output)
	path = strings.ReplaceAll(path, "'", "")
	path = strings.ReplaceAll(path, "\n", "")

	pathNodes := strings.Split(path, string(os.PathSeparator))
	for _, pathNode := range pathNodes {
		rootPath = append(rootPath, pathNode)
	}

	return filepath.Join(rootPath...)
}
func GetProjectRootPath() string {

	cmd := exec.Command("go", "list", "-m", "-f", "'{{.Dir}}'", "github.com/alicenet/alicenet")
	stdout, err := cmd.Output()
	if err != nil {
		panic(fmt.Errorf("Error getting project root path: %v", err))
	}
	return sanitizePathFromOutput(stdout)
}

// GetHardhatPackagePath return the bridge folder path
func GetHardhatPackagePath() string {
	rootPath := GetProjectRootPath()
	bridgePath := filepath.Join(rootPath, SmartContractsRelativeFolder)

	return bridgePath
}

func GenerateHardhatConfig(tempDir string, hardhatPath string, endPoint string) string {
	configTemplate := `
	import "%[1]s/node_modules/@nomiclabs/hardhat-ethers";
	import "%[1]s/node_modules/@nomiclabs/hardhat-truffle5";
	import "%[1]s/node_modules/@nomiclabs/hardhat-waffle";
	import "%[1]s/node_modules/@typechain/hardhat";
	import { HardhatUserConfig} from "%[1]s/node_modules/hardhat/config";
	import "%[1]s/scripts/generateImmutableAuth";
	import "%[1]s/scripts/lib/alicenetFactoryTasks";
	import "%[1]s/scripts/lib/alicenetTasks";
	import "%[1]s/scripts/lib/gogogen";

	const config: HardhatUserConfig = {
		networks: {
			dev: {
			url: "%[2]v",
			accounts: [
				"0x%[3]v",
			],
			},
			hardhat: {
				chainId: 1337,
				allowUnlimitedContractSize: true,
				accounts: [
					{
						privateKey: "0x%[3]v",
						balance: "1500000000000000000000000000000",
					},
				],
			},
		},
		paths: {
			sources: "%[1]s/contracts",
			tests: "%[1]s/test",
			cache: "%[1]s/cache",
			artifacts: "%[1]s/artifacts",
		  },
	};
	export default config;
	`
	hardhatConfigPath := filepath.Join(tempDir, "hardhat.config.ts")
	hardhatConfig, err := os.OpenFile(hardhatConfigPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Errorf("failed to open/create hardhat config file: %v", err))
	}
	data := fmt.Sprintf(
		configTemplate,
		hardhatPath,
		endPoint,
		TestAdminPrivateKey,
	)
	if _, err := hardhatConfig.WriteString(data); err != nil {
		panic(fmt.Errorf("failed to save hardhat config in file: %v", err))
	}

	return hardhatConfigPath
}

// SetCommandStdOut If ENABLE_SCRIPT_LOG env variable is set as 'true' the command will show scripts logs
func SetCommandStdOut(cmd *exec.Cmd) {

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

func executeCommand(dir, command string, args ...string) ([]byte, error) {
	logger := logging.GetLogger("test")
	cmdArgs := strings.Split(strings.Join(args, " "), " ")

	cmd := exec.Command(command, cmdArgs...)
	cmd.Dir = dir
	output, err := cmd.Output()

	if err != nil {
		logger.Errorf("Error executing command: %v %v in dir: %v. %v", command, cmdArgs, dir, string(output))
		return output, err
	}
	logger.Tracef("Command Executed: %v %s in dir: %v. \n%s\n", command, cmdArgs, dir, string(output))

	return output, err
}

type Hardhat struct {
	hostname   string
	port       string
	url        string
	cmd        *exec.Cmd
	configPath string
}

func StartHardHatNodeWithDefaultHost() (*Hardhat, error) {
	return StartHardHatNode("127.0.0.1", "8545")
}

func StartHardHatNode(hostname string, port string) (*Hardhat, error) {
	sanitizedHostname := ""
	sanitizedPort := ""
	if strings.Contains(hostname, "http://") || strings.Contains(hostname, "https://") {
		splits := strings.Split(hostname, "//")
		if len(splits) < 2 {
			return nil, fmt.Errorf("incorrect hostname for hardhat node: %s", hostname)
		}
		sanitizedHostname += splits[1]
	} else {
		sanitizedHostname += hostname
	}
	if port == "" {
		sanitizedPort = "8545"
	} else {
		sanitizedPort = port
	}
	fullUrl := fmt.Sprintf("http://%s:%s", sanitizedHostname, sanitizedPort)

	hardhatTempDir, err := os.MkdirTemp("", "hardhattempdir")
	if err != nil {
		panic(fmt.Errorf("failed to create tmp dir for hardhat: %v", err))
	}

	hardhatPath := GetHardhatPackagePath()
	configPath := GenerateHardhatConfig(hardhatTempDir, hardhatPath, fullUrl)
	hardBinPath := filepath.Join(hardhatPath, "node_modules", ".bin", "hardhat")

	cmd := exec.Command(
		"node",
		hardBinPath,
		"--show-stack-traces",
		"--config",
		configPath,
		"node",
		"--hostname",
		sanitizedHostname,
		"--port",
		sanitizedPort,
	)
	cmd.Dir = GetHardhatPackagePath()

	// setCommandStdOut(cmd)
	err = cmd.Start()
	// if there is an error with our execution handle it here
	if err != nil {
		return nil, fmt.Errorf("could not run hardhat node: %s", err)
	}
	hardhat := &Hardhat{cmd: cmd, url: fullUrl, configPath: configPath, port: port, hostname: hostname}
	ctx, cf := context.WithTimeout(context.Background(), time.Second*100)
	defer cf()
	err = hardhat.WaitForHardHatNode(ctx)
	if err != nil {
		return nil, fmt.Errorf("could wait for hardhat node: %s", err)
	}
	return hardhat, nil
}

func (h *Hardhat) WaitForHardHatNode(ctx context.Context) error {
	logger := logging.GetLogger("test")
	c := http.Client{}
	msg := &ethereum.JsonRPCMessage{
		Version: "2.0",
		ID:      []byte("1"),
		Method:  "eth_chainId",
		Params:  make([]byte, 0),
	}

	params, err := json.Marshal(make([]string, 0))
	if err != nil {
		return fmt.Errorf("could not run hardhat node: %v", err)
	}
	msg.Params = params

	var buff bytes.Buffer
	err = json.NewEncoder(&buff).Encode(msg)
	if err != nil {
		return fmt.Errorf("Error creating a buffer json encoder: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
			body := bytes.NewReader(buff.Bytes())
			_, err := c.Post(
				h.url,
				"application/json",
				body,
			)
			if err != nil {
				continue
			}
			logger.Infof("HardHat node started correctly")
			return nil
		}
	}
}

func (h *Hardhat) Close() error {
	defer os.RemoveAll(h.configPath)

	logger := logging.GetLogger("test")
	logger.Debug("Stopping HardHat running instance")

	process, err := os.FindProcess(h.cmd.Process.Pid)
	if err != nil {
		logger.Errorf("Error finding HardHat pid: %v", err)
		return err
	}

	logger.Debugf("Found hardhat process via pid %d", process.Pid)
	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		logger.Errorf("Error waiting sending SIGTERM signal to HardHat process: %v", err)
		return err
	}
	logger.Debug("Sent sigterm signal to HardHat process, waiting for it to terminate")
	_, err = process.Wait()
	if err != nil {
		logger.Errorf("Error waiting HardHat process to stop: %v", err)
		return err
	}

	logger.Debug("HardHat node has been stopped")
	return nil
}

func (h *Hardhat) IsHardHatRunning() (bool, error) {
	var client = http.Client{Timeout: 5 * time.Second}
	resp, err := client.Head(h.url)
	if err != nil {
		return false, err
	}
	resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return true, nil
	}

	return false, nil
}

func (h *Hardhat) DeployFactoryAndContracts(tmpDir string, baseFilesDir string) (string, error) {
	modulesDir := GetHardhatPackagePath()
	output, err := executeCommand(
		modulesDir,
		"npx",
		"hardhat",
		"--network",
		"dev",
		"--show-stack-traces",
		"--config",
		h.configPath,
		"deployContracts",
		"--input-folder",
		filepath.Join(baseFilesDir),
		"--output-folder",
		tmpDir,
	)
	if err != nil {
		return "", err
	}
	logLines := strings.Split(string(output), "\n")
	factoryAddress := ""
	for _, line := range logLines {
		if strings.Contains(line, "AliceNetFactory") {
			addressLine := strings.Split(line, ":")
			factoryAddress = strings.TrimSpace(addressLine[len(addressLine)-1])
		}
	}
	if factoryAddress == "" {
		return "", fmt.Errorf("unable to find factoryAddress")
	}

	return factoryAddress, nil
}

func (h *Hardhat) RegisterValidators(factoryAddress string, validators []string) error {
	nodeDir := GetHardhatPackagePath()
	// Register validator
	_, err := executeCommand(
		nodeDir,
		"npx",
		"hardhat",
		"--network",
		"dev",
		"--show-stack-traces",
		"--config",
		h.configPath,
		"registerValidators",
		"--test",
		"--factory-address",
		factoryAddress,
		strings.Join(validators, " "),
	)
	if err != nil {
		return err
	}
	return nil
}

// SendCommandViaRPC sends a command to the hardhat server via an RPC call
func SendCommandViaRPC(url string, command string, params ...interface{}) error {
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
		url,
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

// MineBlocks mines a certain number of hardhat blocks
func MineBlocks(endPoint string, blocksToMine uint64) {
	logger := logging.GetLogger("test")
	var blocksToMineString = "0x" + strconv.FormatUint(blocksToMine, 16)
	logger.Tracef("hardhat_mine %v blocks ", blocksToMine)
	err := SendCommandViaRPC(endPoint, "hardhat_mine", blocksToMineString)
	if err != nil {
		panic(err)
	}
}

// Mine finality delay blocks + 1
func MineFinalityDelayBlocks(client layer1.Client) {
	MineBlocks(client.GetEndpoint(), client.GetFinalityDelay()+1)
}

// AdvanceTo advances to a certain block number
func AdvanceTo(eth layer1.Client, target uint64) {
	logger := logging.GetLogger("test")
	currentBlock, err := eth.GetCurrentHeight(context.Background())
	if err != nil {
		panic(err)
	}
	if target < currentBlock {
		return
	}
	blocksToMine := target - currentBlock
	var blocksToMineString = "0x" + strconv.FormatUint(blocksToMine, 16)

	logger.Tracef("hardhat_mine %v blocks to target height %v", blocksToMine, target)

	err = SendCommandViaRPC(eth.GetEndpoint(), "hardhat_mine", blocksToMineString)
	if err != nil {
		panic(err)
	}
}

// SetNextBlockBaseFee sets the the Base fee for the next hardhat block. Can be used to make tx stale.
func SetNextBlockBaseFee(endPoint string, target uint64) {
	logger := logging.GetLogger("test")
	logger.Tracef("Setting hardhat_setNextBlockBaseFeePerGas to %v", target)
	err := SendCommandViaRPC(endPoint, "hardhat_setNextBlockBaseFeePerGas", "0x"+strconv.FormatUint(target, 16))
	if err != nil {
		panic(err)
	}
}

// SetAutoMine enables/disables hardhat autoMine
func SetAutoMine(endPoint string, autoMine bool) {
	logger := logging.GetLogger("test")
	logger.Tracef("Setting Automine to %v", autoMine)

	err := SendCommandViaRPC(endPoint, "evm_setAutomine", autoMine)
	if err != nil {
		panic(err)
	}
}

// ResetHardhatNode resets hardhat node from scratch
func ResetHardhatNode(endPoint string) {
	logger := logging.GetLogger("test")
	logger.Trace("Resetting hardhat node from scratch")

	err := SendCommandViaRPC(endPoint, "hardhat_reset")
	if err != nil {
		panic(err)
	}
}

// SetBlockInterval sets the interval between hardhat blocks. In case interval is 0, we enter in
// manual mode and blocks can only be mined explicitly by calling `MineBlocks`.
// This function disables autoMine.
func SetBlockInterval(endPoint string, intervalInMilliSeconds uint64) {
	SetAutoMine(endPoint, false)
	logger := logging.GetLogger("test")
	logger.Tracef("Setting block interval to %v seconds", intervalInMilliSeconds)
	err := SendCommandViaRPC(endPoint, "evm_setIntervalMining", intervalInMilliSeconds)
	if err != nil {
		panic(err)
	}
}

// ResetHardhatConfigs resets the hardhat configs to automine true and basefee 100GWei
func ResetHardhatConfigs(endPoint string) {
	ResetHardhatNode(endPoint)
	SetAutoMine(endPoint, true)
	SetNextBlockBaseFee(endPoint, 100_000_000_000)
}
