package main

import (
	"fmt"
	"time"

	"github.com/MadBase/MadNet/ipc"
)

func main() {
	s := ipc.NewServer("/tmp/madnet_firewalld")

	go func() {
		err := s.Start()
		if err != nil {
			panic(err)
		}
	}()

	var updates [][]string = [][]string{
		{"11.22.33.44:55"},
		{"11.22.33.44:55", "22.33.44.55:66"},
		{"11.22.33.44:55", "22.33.44.55:66", "33.44.55.66:77"},
		{"22.33.44.55:66", "33.44.55.66:77"},
		{"33.44.55.66:77"},
		{"33.44.55.66:77", "44.55.66.77:88"},
		{"33.44.55.66:77", "44.55.66.77:88", "55.66.77.88:99"},
		{"44.55.66.77:88", "55.66.77.88:99"},
		{"55.66.77.88:99"},
	}

	for i, v := range updates {
		time.Sleep(time.Second * 15)
		err := s.Push(ipc.PeersUpdate{Addrs: v, Seq: uint(i)})
		if err != nil {
			fmt.Printf("push err: %T %v\n", err, err)
		}
	}

}
