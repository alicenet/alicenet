package firewalld

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
)

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

func getAllowedIPs(ruleName string) (IPset, error) {
	res, err := runCmd("gcloud", "compute", "firewall-rules", "describe", ruleName, "--format", "json")
	if err != nil {
		return nil, err
	}

	var resp struct {
		SourceRanges []string `json:"sourceRanges"`
	}
	err = json.Unmarshal(res, &resp)
	if err != nil {
		return nil, err
	}

	return NewIPset(resp.SourceRanges), nil
}

func setAllowedIPs(ruleName string, allowedIPs IPset) error {
	_, err := runCmd("gcloud", "compute", "firewall-rules", "update", ruleName, "--source-ranges", allowedIPs.MarshallString())
	return err
}
