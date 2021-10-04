package gcloud

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

func getRulePrefix() (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1/instance/id", nil)
	if err != nil {
		return "", nil
	}

	req.Header.Add("Metadata-Flavor", `Google`)
	resp, err := client.Do(req)
	if err != nil {
		return "", nil
	}
	defer resp.Body.Close()

	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil
	}

	instanceId := string(res)
	return "firewalld-" + instanceId + "-", nil
}

func createRuleName(rulePrefix string, ip string, port string) string {
	return rulePrefix + strings.ReplaceAll(ip, ".", "-") + "--" + port
}

func splitRuleName(rulePrefix string, ruleName string) (string, error) {
	if !strings.HasPrefix(ruleName, rulePrefix) {
		return "", errors.Errorf("rule %v does not have prefix %v", ruleName, rulePrefix)
	}
	address := ruleName[len(rulePrefix)+1:]

	addressParts := strings.SplitN(address, "--", 2)
	if len(addressParts) != 2 {
		return "", errors.Errorf("rule %v does not have --", ruleName)
	}

	return strings.ReplaceAll(addressParts[0], "-", ".") + ":" + addressParts[1], nil
}
