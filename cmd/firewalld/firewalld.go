package firewalld

import (
	"encoding/json"
	"io"
	"net"
	"time"

	"github.com/MadBase/MadNet/cmd/firewalld/gcloud"
	"github.com/MadBase/MadNet/cmd/firewalld/lib"
	"github.com/MadBase/MadNet/config"
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

var Command = cobra.Command{
	Use:   "firewalld",
	Short: "Continously updates a given firewall in the background",
	Long:  "Continously updates a given firewall using values received over IPC from main MadNet process",
	Run:   FirewallDaemon,
}

var implementations = map[string]lib.ImplementationConstructor{
	"gcloud": gcloud.NewImplementation,
}

func FirewallDaemon(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		panic("must provide firewall implementation parameter")
	}

	implementationStr := args[0]
	constructor := implementations[implementationStr]
	if constructor == nil {
		panic("invalid firewall implementation paramater")
	}

	socketFile := config.Configuration.Firewalld.SocketFile
	if socketFile == "" {
		panic("must have config option firewalld.socketfile set")
	}

	logger := logging.GetLogger(constants.LoggerFirewalld)

	im, err := constructor(logger)
	if err != nil {
		panic(err)
	}

	for {
		startConnection(logger, socketFile, im)
	}
}

func startConnection(logger *logrus.Logger, socketFile string, im lib.Implementation) {
	l, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: socketFile, Net: "unix"})
	if err != nil {
		logger.Info("Connection failed (retry in 5s..): ", err)
		time.Sleep(5 * time.Second)
		return
	}

	defer func() {
		l.Close()
		logger.Info("Disconnected!")
	}()

	logger.Info("Connected!")

	_, err = l.Write([]byte(`{"jsonrpc":"2.0","id":"sub","method":"subscribe"}`))
	logger.Debug("Subscription requested!")

	if err != nil {
		panic(err)
	}

	dec := json.NewDecoder(l)
	var r Response
	for {
		err := dec.Decode(&r)

		if err == io.EOF {
			logger.Debug("Received EOF!")
			break
		}
		if err != nil {
			logger.Error("Error during receive: ", err, r)
			break
		}
		if r.Error != nil {
			logger.Error("Error object returned: ", r.Error)
			continue
		}
		if r.Id != "sub" {
			logger.Warn("Message with unknown id: ", r.Id, ", skipping...")
			continue
		}

		var result SubscriptionResult
		err = json.Unmarshal(r.Result, &result)
		if err != nil {
			logger.Error("Error unmarshalling result: ", err)
			continue
		}

		currentAddrs, err := im.GetAllowedAddresses()
		if err != nil {
			logger.Error("Failed to get firewall: ", err)
			break
		}
		desiredAddrs := lib.NewAddresSet(result.Addrs)

		toDelete := lib.AddressSet{}
		for a := range currentAddrs {
			if !desiredAddrs.Has(a) {
				toDelete.Add(a)
			}
		}

		toAdd := lib.AddressSet{}
		for a := range desiredAddrs {
			if !currentAddrs.Has(a) {
				toAdd.Add(a)
			}
		}

		logger.Debugf("Updating firewall, adding %v, deleting %v...", toAdd, toDelete)
		err = im.UpdateAllowedAddresses(toAdd, toDelete)
		if err != nil {
			logger.Error("Failed to update firewall", err)
			break
		}

		logger.Info("Updated firewall successfully!")
	}
}
