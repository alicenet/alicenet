package firewalld

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

type ErrUpdate struct {
	msg     string
	outputs []error
}

func (e ErrUpdate) Error() string {
	ret := e.msg + ":"
	for _, v := range e.outputs {
		ret += "\n" + v.Error()
	}
	return ret
}

func runCmd(c ...string) ([]byte, error) {
	cmd := exec.Command(c[0], c[1:]...)
	pipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	defer pipe.Close()

	var stderr []byte
	go func() {
		stderr, _ = ioutil.ReadAll(pipe)
	}()
	o, err := cmd.Output()

	if err != nil {
		return nil, fmt.Errorf("%v: %v", err, string(stderr))
	}

	return o, nil
}

func updateRules(logger *logrus.Logger, rulePrefix string, desired AddressSet) error {
	res, err := runCmd("gcloud", "compute", "firewall-rules", "list", "--filter", "name~"+rulePrefix, "--format", "value(name)")
	if err != nil {
		return ErrUpdate{"Could not find rules", []error{err}}
	}

	start := 0
	current := AddressSet{}
	for i := range res {
		if res[i] != '\n' {
			continue
		}
		current.Add(splitRuleName(rulePrefix, string(res[start:i])))
		start = i + 1
	}

	count := 0
	errs := make(chan error)
	for a := range desired {
		if !current.Has(a) {
			logger.Debug("adding", a)
			go func(addr string) { errs <- addRule(rulePrefix, addr) }(a)
			count++
		}
	}
	for a := range current {
		if !desired.Has(a) {
			logger.Debug("deleting", a)
			go func(addr string) { errs <- deleteRule(rulePrefix, addr) }(a)
			count++
		}
	}

	errOutputs := make([]error, 0)
	for i := 0; i < count; i++ {
		e := <-errs
		if e != nil {
			errOutputs = append(errOutputs, e)
		}
	}

	if len(errOutputs) > 0 {
		return ErrUpdate{"Some create/delete commands failed to run", errOutputs}
	}
	return nil
}

func addRule(rulePrefix string, addr string) error {
	addrParts := strings.SplitN(addr, ":", 2)
	_, err := runCmd("gcloud", "compute", "firewall-rules", "create", createRuleName(rulePrefix, addrParts[0], addrParts[1]), "--source-ranges", addrParts[0], "--allow", "tcp:"+addrParts[1])
	return err
}

func deleteRule(rulePrefix string, addr string) error {
	addrParts := strings.SplitN(addr, ":", 2)
	_, err := runCmd("gcloud", "compute", "firewall-rules", "delete", createRuleName(rulePrefix, addrParts[0], addrParts[1]))
	return err
}
