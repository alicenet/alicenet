package gcloud

import (
	"fmt"
	"strings"
)

func createRulePrefix(instanceId string) string {
	return "firewalld-" + instanceId
}

func createRuleName(rulePrefix string, ip string, port string) string {
	return rulePrefix + "-" + strings.ReplaceAll(ip, ".", "-") + "--" + port
}

func splitRuleName(rulePrefix string, ruleName string) (string, error) {
	if !strings.HasPrefix(ruleName, rulePrefix) {
		return "", fmt.Errorf("rule %v does not have prefix %v", ruleName, rulePrefix)
	}
	address := ruleName[len(rulePrefix)+1:]

	addressParts := strings.SplitN(address, "--", 2)
	if len(addressParts) != 2 {
		return "", fmt.Errorf("rule %v does not have --", ruleName)
	}

	return strings.ReplaceAll(addressParts[0], "-", ".") + ":" + addressParts[1], nil
}
