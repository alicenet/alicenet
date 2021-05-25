package tasks_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFoo(t *testing.T) {
	var s map[string]string

	raw, err := json.Marshal(s)
	assert.Nil(t, err)

	t.Logf("Raw data:%v", string(raw))
}
