package firewalld

import (
	"io/ioutil"
	"net/http"
	"strings"
)

func getInstanceId() (string, error) {
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

	return string(res), nil
}

const prefix1 = "firewalld"

func getRulePrefix(instanceId string) string {
	return prefix1 + "-" + instanceId
}

func createRuleName(rulePrefix string, ip string, port string) string {
	return rulePrefix + "-" + strings.ReplaceAll(ip, ".", "-") + "--" + port
}

func splitRuleName(rulePrefix string, ruleName string) string {
	address := ruleName[len(rulePrefix)+1:]
	addressParts := strings.SplitN(address, "--", 2)
	return strings.ReplaceAll(addressParts[0], "-", ".") + ":" + addressParts[1]
}
