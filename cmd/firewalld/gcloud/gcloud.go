package gcloud

import (
	"strings"

	"github.com/alicenet/alicenet/cmd/firewalld/lib"
	"github.com/sirupsen/logrus"
)

type Implementation struct {
	rulePrefix string
	runCmd     func(c ...string) ([]byte, error)
	logger     *logrus.Logger
}

func NewImplementation(logger *logrus.Logger) (lib.Implementation, error) {
	instanceId, err := getInstanceId()
	if err != nil {
		return nil, err
	}
	prefix := createRulePrefix(instanceId)

	zone, err := getZone()
	if err != nil {
		return nil, err
	}
	// add unique tag to self (command is noop if tag already exists)
	_, err = lib.RunCmd("gcloud", "-q", "compute", "instances", "add-tags", instanceId, "--zone", zone, "--tags", prefix)
	if err != nil {
		return nil, err
	}

	return &Implementation{prefix, lib.RunCmd, logger}, nil
}

func (im *Implementation) GetAllowedAddresses() (lib.AddressSet, error) {
	res, err := im.runCmd("gcloud", "-q", "compute", "firewall-rules", "list", "--filter", "name~"+im.rulePrefix, "--format", "value(name)")
	if err != nil {
		return nil, lib.ErrCmd{Msg: "could not find rules", Outputs: []error{err}}
	}

	start := 0
	allowed := lib.AddressSet{}
	for i := range res {
		if res[i] != '\n' {
			continue
		}
		addr, err := splitRuleName(im.rulePrefix, string(res[start:i]))
		if err != nil {
			return nil, err
		}
		allowed.Add(addr)
		start = i + 1
	}

	return allowed, nil
}

func (im *Implementation) UpdateAllowedAddresses(toAdd lib.AddressSet, toDelete lib.AddressSet) error {
	count := 0
	errs := make(chan error)
	for a := range toAdd {
		go func(addr string) { errs <- im.addRule(addr) }(a)
		count++
	}
	for a := range toDelete {
		go func(addr string) { errs <- im.deleteRule(addr) }(a)
		count++
	}

	errOutputs := make([]error, 0)
	for i := 0; i < count; i++ {
		e := <-errs
		if e != nil {
			errOutputs = append(errOutputs, e)
		}
	}

	if len(errOutputs) > 0 {
		return lib.ErrCmd{Msg: "Some create/delete commands failed to run", Outputs: errOutputs}
	}
	return nil
}

func (im *Implementation) addRule(addr string) error {
	addrParts := strings.SplitN(addr, ":", 2)
	cmd := []string{"gcloud", "-q", "compute", "firewall-rules", "create", createRuleName(im.rulePrefix, addrParts[0], addrParts[1]), "--target-tags", im.rulePrefix, "--source-ranges", addrParts[0], "--allow", "tcp:" + addrParts[1]}
	im.logger.Tracef("Running command: %v", cmd)
	_, err := im.runCmd(cmd...)
	return err
}

func (im *Implementation) deleteRule(addr string) error {
	addrParts := strings.SplitN(addr, ":", 2)
	cmd := []string{"gcloud", "-q", "compute", "firewall-rules", "delete", createRuleName(im.rulePrefix, addrParts[0], addrParts[1])}
	im.logger.Tracef("Running command: %v", cmd)
	_, err := im.runCmd(cmd...)
	return err
}

var _ lib.ImplementationConstructor = NewImplementation
