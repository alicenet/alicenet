package ethkey

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_GetPassword(t *testing.T) {
	pass1 := getPassword()
	fmt.Println(pass1)

	pass2 := getPassword()
	fmt.Println(pass2)

	pass3 := getPassword()
	fmt.Println(pass3)

	assert.NotEqual(t, pass1, pass2)
	assert.NotEqual(t, pass1, pass3)
	assert.NotEqual(t, pass2, pass3)
}
