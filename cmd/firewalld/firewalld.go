package firewalld

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/alicenet/alicenet/cmd/firewalld/gcloud"
	"github.com/alicenet/alicenet/cmd/firewalld/lib"
	"github.com/alicenet/alicenet/config"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/logging"
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

var ErrNoConn = fmt.Errorf("no conn")

var Command = cobra.Command{
	Use:   "firewalld",
	Short: "Continously updates a given firewall in the background",
	Long:  "Continously updates a given firewall using values received over IPC from main AliceNet process",
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
		err := startConnection(logger, socketFile, im)
		if err == ErrNoConn {
			logger.Info("Connection failed (retry in 5s..): ", err)
			time.Sleep(5 * time.Second)
		}
	}
}

func startConnection(logger *logrus.Logger, socketFile string, im lib.Implementation) error {
	l, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: socketFile, Net: "unix"})
	if err != nil {
		return ErrNoConn
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
			return io.EOF
		}
		if err != nil {
			logger.Error("Error during receive: ", err, r)
			return err
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
			return err
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
			return err
		}

		logger.Info("Updated firewall successfully!")
	}
}
