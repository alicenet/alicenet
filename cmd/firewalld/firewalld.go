package firewalld

import (
	"fmt"

	"encoding/json"
	"io"
	"net"
	"time"

	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type SubscriptionResult struct {
	Addrs []string
	Seq   uint
}

type Response struct {
	Id     string          `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *struct {
		code    int
		message string
		data    interface{}
	} `json:"error,omitempty"`
}

const socketFile = "/tmp/foobar"

var Command = cobra.Command{
	Use:   "firewalld",
	Short: "Continously updates a given gcp firewall in the background",
	Long:  "Continously updates a given gcp firewall using values received over IPC from main MadNet process",
	Run:   FirewallDaemon}

func FirewallDaemon(cmd *cobra.Command, args []string) {
	instanceId, err := getInstanceId()
	if err != nil {
		panic(fmt.Errorf("Could not retrieve instanceId: %v", err))
	}

	prefix := getRulePrefix(instanceId)
	logger := logging.GetLogger(constants.LoggerFirewalld)

	for {
		startConnection(logger, prefix)
	}
}

func startConnection(logger *logrus.Logger, rulePrefix string) {
	l, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: socketFile, Net: "unix"})
	if err != nil {
		logger.Info("Connection failed (retry in 5s..):", err)
		time.Sleep(5 * time.Second)
		return
	}

	defer func() {
		l.Close()
		logger.Info("Disconnected!")
	}()

	logger.Info("Connected!")

	_, err = l.Write([]byte(`{"jsonrpc":"2.0","id":"sub","method":"subscribe"}`))
	logger.Info("Subscription requested!")

	if err != nil {
		panic(err)
	}

	dec := json.NewDecoder(l)
	var r Response
	for {
		err := dec.Decode(&r)

		if err == io.EOF {
			logger.Info("Received EOF!")
			break
		}
		if err != nil {
			logger.Info("Error during receive: ", err, r)
			break
		}
		if r.Error != nil {
			logger.Info("Error object returned", r.Error)
			continue
		}
		if r.Id != "sub" {
			logger.Info("Message with unknown id:", r.Id, ", skipping...")
			continue
		}

		var result SubscriptionResult
		err = json.Unmarshal(r.Result, &result)
		if err != nil {
			logger.Info("Error unmarshalling result", err)
			continue
		}

		logger.Info("done updating rules", updateRules(logger, rulePrefix, NewAddresSet(result.Addrs)))
	}
}
