package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"
)

type Address struct {
	IP   [4]byte
	Port uint16
}

type Update struct {
	Add    []Address
	Delete []Address
}

const file = "/tmp/foobar"

func main() {

	for {
		func() {
			l, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: file, Net: "unix"})
			if err != nil {
				fmt.Println("CONN ERR:", err)
				time.Sleep(time.Second)
				return
			}
			defer l.Close()

			fmt.Println("CONNECTED!")

			_, err = l.Write([]byte(`[{"jsonrpc":"2.0","id":"h","method":"healthz"},{"jsonrpc":"2.0","id":"sub","method":"subscribe"}]`))
			if err != nil {
				panic(err)
			}

			time.Sleep(time.Second * 3)
			dec := json.NewDecoder(l)
			var u interface{}
			for {

				err := dec.Decode(&u)

				if err == io.EOF {
					fmt.Println("DISCONNECTED!")
					break
				} else if err != nil {
					fmt.Println("ERR: ", err)
					break
				} else {
					fmt.Println("DATA:", u)
				}
			}
		}()
	}

}
