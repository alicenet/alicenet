package lib

import (
	"fmt"
	"io"
	"os/exec"
)

type ErrCmd struct {
	Msg     string
	Outputs []error
}

func (e ErrCmd) Error() string {
	ret := e.Msg + ":"
	for _, v := range e.Outputs {
		ret += "\n" + v.Error()
	}
	return ret
}

func RunCmd(c ...string) ([]byte, error) {
	cmd := exec.Command(c[0], c[1:]...)
	pipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	defer pipe.Close()

	var stderr []byte
	go func() {
		stderr, _ = io.ReadAll(pipe)
	}()
	o, err := cmd.Output()

	if err != nil {
		return nil, fmt.Errorf("%v: %v", err, string(stderr))
	}

	return o, nil
}
