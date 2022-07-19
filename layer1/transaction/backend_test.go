package transaction

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInfoSaveAndLoad(t *testing.T) {
	originalMap := map[FuncSelector]string{FuncSelector{1, 1, 1, 1}: "selector"}
	originalMapBytes, err := json.Marshal(originalMap)
	assert.Nil(t, err)
	fmt.Printf("%v\n", originalMap)

	resultMap := map[FuncSelector]string{}
	err = json.Unmarshal(originalMapBytes, &resultMap)
	assert.Nil(t, err)
	fmt.Printf("%v\n", resultMap)
	assert.Equal(t, originalMap, resultMap)

	originalInfo := &info{
		Selector: &FuncSelector{2, 2, 2, 2},
	}
	fmt.Printf("%v\n", originalInfo.Selector)

	originalInfoBytes, err := json.Marshal(originalInfo)
	assert.Nil(t, err)

	resultInfo := &info{}
	err = json.Unmarshal(originalInfoBytes, resultInfo)
	assert.Nil(t, err)
	fmt.Printf("%v\n", resultInfo.Selector)
	assert.Equal(t, originalInfo, resultInfo)
}
