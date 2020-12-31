package utxotrie

import (
	"github.com/MadBase/MadNet/constants/dbprefix"
)

func getTriePrefix() []byte {
	return dbprefix.PrefixUTXOTrie()
}
