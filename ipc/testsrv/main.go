package main

import (
	"fmt"
	"time"

	"github.com/MadBase/MadNet/ipc"
)

func main() {
	s := ipc.NewServer("/tmp/foobar")

	exited := false
	go func() {
		time.Sleep(time.Second * 20)
		s.Close()
		fmt.Println("called close")
		exited = true
	}()

	err := s.Start()
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Second * 2)

	var i uint64 = 0
	for {
		time.Sleep(time.Second)
		err := s.Push(ipc.PeersUpdate{Add: []ipc.Address{{IP: [...]byte{byte(i), byte(i), byte(i), byte(i)}, Port: uint16(i)}}, Delete: []ipc.Address{}})
		if err != nil {
			fmt.Printf("push err: %T %v\n", err, err)
			if exited {
				break
			}
		}
		i++
	}

}
