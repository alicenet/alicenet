package gcloud

import (
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/MadBase/MadNet/cmd/firewalld/lib"
	"github.com/sirupsen/logrus/hooks/test"
)

type mockCmder struct {
	in       [][]string
	outBytes []byte
	outErr   error
	mu       sync.Mutex
}

func newMockCmder(outBytes []byte, outErr error) *mockCmder {
	m := mockCmder{in: [][]string{}, outBytes: outBytes, outErr: outErr}
	return &m
}

func (m *mockCmder) RunCmd(c ...string) ([]byte, error) {
	m.mu.Lock()
	m.in = append(m.in, c)
	m.mu.Unlock()

	return m.outBytes, m.outErr
}

func (m *mockCmder) Called(c ...string) bool {
	for _, v := range m.in {
		if reflect.DeepEqual(v, c) {
			return true
		}
	}
	return false
}

var logger, _ = test.NewNullLogger()

func TestGetAllowedAddresses(t *testing.T) {
	m := newMockCmder([]byte("testprefix-12-23-34-45--5678\ntestprefix-11-22-33-44--5555\n"), nil)
	c := &Implementation{"testprefix-", m.RunCmd, logger}
	b, err := c.GetAllowedAddresses()

	if err != nil {
		t.Fatal("GetAllowedAddresses returned error ", err)
	}
	if len(b) != 2 || !b["2.23.34.45:5678"] || !b["1.22.33.44:5555"] {
		t.Fatal("GetAllowedAddresses returned incorrect results ", b)
	}
	if len(m.in) != 1 || !m.Called("gcloud", "compute", "firewall-rules", "list", "--filter", "name~testprefix-", "--format", "value(name)") {
		t.Fatalf("Command run was not the expected command: %v", m.in)
	}
}

func TestGetAllowedAddressesInvalid(t *testing.T) {
	m := newMockCmder([]byte("testttttprefix-12-23-34-45--5678\ntestprefix-11-22-33-44--5555\n"), nil)
	c := &Implementation{"testprefix-", m.RunCmd, logger}
	_, err := c.GetAllowedAddresses()

	if err == nil {
		t.Fatal("Should throw error", err)
	}
}

func TestGetAllowedAddressesError(t *testing.T) {
	m := newMockCmder([]byte(""), fmt.Errorf("Nope"))
	c := &Implementation{"testprefix-", m.RunCmd, logger}
	_, err := c.GetAllowedAddresses()

	if err == nil {
		t.Fatal("Should throw error", err)
	}
}

func TestGetAllowedAddressesEmpty(t *testing.T) {
	m := newMockCmder([]byte(""), nil)
	c := &Implementation{"testprefix-", m.RunCmd, logger}
	_, err := c.GetAllowedAddresses()

	if err != nil {
		t.Fatal("Should not throw error", err)
	}
}

func TestUpdateAllowedAddresses(t *testing.T) {
	m := newMockCmder([]byte(""), nil)
	c := &Implementation{"testprefix-", m.RunCmd, logger}

	err := c.UpdateAllowedAddresses(
		lib.NewAddresSet([]string{"11.22.33.44:5678", "22.33.44.55:6789"}),
		lib.NewAddresSet([]string{"33.44.55.66:7890"}),
	)

	if err != nil {
		t.Fatal("Should not throw error", err)
	}
	if len(m.in) != 3 ||
		!m.Called("gcloud", "compute", "firewall-rules", "create", "testprefix-11-22-33-44--5678", "--source-ranges", "11.22.33.44", "--allow", "tcp:5678") ||
		!m.Called("gcloud", "compute", "firewall-rules", "create", "testprefix-22-33-44-55--6789", "--source-ranges", "22.33.44.55", "--allow", "tcp:6789") ||
		!m.Called("gcloud", "compute", "firewall-rules", "delete", "testprefix-33-44-55-66--7890") {
		t.Fatalf("Commands run were not the expected commands: %v", m.in)
	}
}

func TestUpdateAllowedAddressesError(t *testing.T) {
	m := newMockCmder([]byte(""), fmt.Errorf("oh noes!"))
	c := &Implementation{"testprefix-", m.RunCmd, logger}

	err := c.UpdateAllowedAddresses(
		lib.NewAddresSet([]string{"11.22.33.44:5678", "22.33.44.55:6789"}),
		lib.NewAddresSet([]string{"33.44.55.66:7890"}),
	)

	if err == nil {
		t.Fatal("Should throw error", err)
	}
}

func TestUpdateAllowedAddressesEmpty(t *testing.T) {
	m := newMockCmder([]byte(""), nil)
	c := &Implementation{"testprefix-", m.RunCmd, logger}

	err := c.UpdateAllowedAddresses(
		lib.AddressSet{},
		lib.AddressSet{},
	)

	if err != nil {
		t.Fatal("Should not throw error", err)
	}
	if len(m.in) != 0 {
		t.Fatalf("Should not run commands: %v", m.in)
	}
}
