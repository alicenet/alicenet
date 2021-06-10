package blockchain

import (
	"sync"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/ethereum/go-ethereum/crypto"
)

type SelectorMapDetail struct {
	sync.RWMutex
	signatures map[interfaces.FuncSelector]string
	selectors  map[string]interfaces.FuncSelector
}

func NewSelectorMap() *SelectorMapDetail {
	return &SelectorMapDetail{
		signatures: make(map[interfaces.FuncSelector]string, 20),
		selectors:  make(map[string]interfaces.FuncSelector, 20)}
}

func (selectorMap *SelectorMapDetail) Selector(signature string) interfaces.FuncSelector {

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

func (selectorMap *SelectorMapDetail) Signature(selector interfaces.FuncSelector) string {
	selectorMap.RLock()
	defer selectorMap.RUnlock()

	return selectorMap.signatures[selector]
}

// CalculateSelector calculates the hash of the supplied function signature
func CalculateSelector(signature string) interfaces.FuncSelector {
	var selector [4]byte

	selectorSlice := crypto.Keccak256([]byte(signature))[:4]
	selector[0] = selectorSlice[0]
	selector[1] = selectorSlice[1]
	selector[2] = selectorSlice[2]
	selector[3] = selectorSlice[3]

	return selector
}

func ExtractSelector(data []byte) interfaces.FuncSelector {
	var selector [4]byte

	if len(data) >= 4 {
		for idx := 0; idx < 4; idx++ {
			selector[idx] = data[idx]
		}
	}

	return selector
}
