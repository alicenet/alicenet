package firewalld

import (
	"fmt"

	"encoding/json"
	"io"
	"net"
	"time"

	"github.com/spf13/cobra"
)

type SubscriptionResult struct {
	Add    []string
	Delete []string
}

type Response struct {
	JsonRpc string          `json:"jsonrpc"`
	Id      string          `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *struct {
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
	if len(args) < 1 {
		panic("no firewall rule name passed")
	}
	rule := args[0]

	allowedIPs, err := getAllowedIPs(rule)
	if err != nil {
		panic(fmt.Errorf("could not find rule %v: %T %v", rule, err, err))
	}

	for {
		startConnection(rule, allowedIPs)
	}
}

func startConnection(rule string, allowedIPs IPset) {
	l, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: socketFile, Net: "unix"})
	if err != nil {
		fmt.Println("Connection failed (retry in 1s..):", err)
		time.Sleep(time.Second)
		return
	}

	defer func() {
		l.Close()
		fmt.Println("Disconnected!")
	}()

	fmt.Println("Connected!")

	_, err = l.Write([]byte(`{"jsonrpc":"2.0","id":"sub","method":"subscribe"}`))
	fmt.Println("Subscription requested!")

	if err != nil {
		panic(err)
	}

	dec := json.NewDecoder(l)
	var r Response
	for {
		err := dec.Decode(&r)

		if err == io.EOF {
			fmt.Println("Received EOF!")
			break
		}
		if err != nil {
			fmt.Println("Error during receive: ", err, r)
			break
		}
		if r.Error != nil {
			fmt.Println("Error object returned", r.Error)
			continue
		}
		if r.Id != "sub" {
			fmt.Println("Message with unknown id:", r.Id, ", skipping...")
			continue
		}

		var s SubscriptionResult
		err = json.Unmarshal(r.Result, &s)
		if err != nil {
			fmt.Println("Error unmarshalling result", err)
			continue
		}

		for _, ip := range s.Add {
			allowedIPs.Add(ip)
		}

		for _, ip := range s.Delete {
			allowedIPs.Delete(ip)
		}

		setAllowedIPs(rule, allowedIPs)
	}
}
