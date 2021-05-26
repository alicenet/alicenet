package blockchain

import (
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
)

type FuncSelector [4]byte

type selectorMaps struct {
	sync.RWMutex
	signatures map[FuncSelector]string
	selectors  map[string]FuncSelector
}

selectorMap := selectorMaps {signatures: map[FuncSelector]string{}, selectors: map[string]FuncSelector{}}

func Selector(signature string) FuncSelector {

	// First check if we already have it
	selectorMap.RLock()
	selector, present := selectorMap.selectors[signature]
	selectorMap.RUnlock()
	if present {
		return selector
	}

	// Calculate and store value
	selector = CalculateSelector(signature)

	selectorMap.Lock()
	selectorMap.signatures[selector] = signature
	selectorMap.Unlock()

	return selector
}

func Signature(selector FuncSelector) string {
	selectorMap.RLock()
	defer selectorMap.RUnlock()

	return selectorMap.signatures[selector]
}

// CalculateSelector calculates the hash of the supplied function signature
func CalculateSelector(signature string) [4]byte {
	var selector [4]byte

	selectorSlice := crypto.Keccak256([]byte(signature))[:4]
	selector[0] = selectorSlice[0]
	selector[1] = selectorSlice[1]
	selector[2] = selectorSlice[2]
	selector[3] = selectorSlice[3]

	return selector
}
