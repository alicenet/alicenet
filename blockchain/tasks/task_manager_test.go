package tasks_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/MadBase/MadNet/blockchain/dkg/dkgtasks"
	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/stretchr/testify/assert"
)

func TestFoo(t *testing.T) {
	var s map[string]string

	raw, err := json.Marshal(s)
	assert.Nil(t, err)

	t.Logf("Raw data:%v", string(raw))
}

func TestType(t *testing.T) {
	state := objects.NewDkgState(accounts.Account{})
	ct := dkgtasks.NewCompletionTask(state)

	var task interfaces.Task = ct
	raw, err := json.Marshal(task)
	assert.Nil(t, err)
	assert.Greater(t, len(raw), 1)

	tipe := reflect.TypeOf(task)
	t.Logf("type0:%v", tipe.String())

	if tipe.Kind() == reflect.Ptr {
		tipe = tipe.Elem()
	}

	typeName := tipe.String()
	t.Logf("type1:%v", typeName)

}
