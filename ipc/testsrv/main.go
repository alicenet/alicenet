package main

import (
	"fmt"
	"time"

	"github.com/MadBase/MadNet/ipc"
)

func main() {
	s := ipc.NewServer("/tmp/foobar")

	go func() {
		err := s.Start()
		if err != nil {
			panic(err)
		}
	}()

	var i uint64 = 0
	for {
		time.Sleep(time.Second)
		b := byte(i)
		p := int16(i)
		err := s.Push(ipc.PeersUpdate{Add: []string{string(b) + "." + string(b) + "." + string(b) + "." + string(b) + ":" + string(p)}, Delete: []string{}})
		if err != nil {
			fmt.Printf("push err: %T %v\n", err, err)
		}
		i++
	}

}
