package utxotrie

import (
	"github.com/alicenet/alicenet/constants/dbprefix"
)

func getTriePrefix() []byte {
	return dbprefix.PrefixUTXOTrie()
}
