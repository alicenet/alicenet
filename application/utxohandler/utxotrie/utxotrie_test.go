package utxotrie

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/test/mocks"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestDataUpdates struct {
	Updates []TestDataUpdate `json:"updates"`
}

type TestDataUpdate struct {
	Keys   []string `json:"keys"`
	Values []string `json:"values"`
}

func TestCorruptedSparseMerkleTrie(t *testing.T) {
	testCases := []struct {
		name       string
		goldenFile string
	}{
		{
			name:       "Simplified",
			goldenFile: "corrupted_smt_simplified.json",
		},
		{
			name:       "Full",
			goldenFile: "corrupted_smt_full.json",
		},
	}

	for _, testCase := range testCases {
		func() {
			db := mocks.NewTestDB()
			trie := NewUTXOTrie(db.DB())

			goldenFile, err := os.Open("./testdata/" + testCase.goldenFile)
			require.Nil(t, err)
			defer goldenFile.Close()

			byteGoldenFile, err := os.ReadFile(goldenFile.Name())
			require.Nil(t, err)

			testData := &TestDataUpdates{}
			err = json.Unmarshal(byteGoldenFile, testData)
			require.Nil(t, err)

			height := 1
			err = db.Update(func(txn *badger.Txn) error {
				_, err := trie.ApplyState(txn, objs.TxVec{}, uint32(height))
				height++
				return err
			})
			require.Nil(t, err)

			err = db.Update(func(txn *badger.Txn) error {
				for _, update := range testData.Updates {
					if len(update.Keys) != len(update.Values) {
						t.Fatal("the keys and value length must be the same")
					}

					addkeys := [][]byte{}
					addvalues := [][]byte{}

					for i := 0; i < len(update.Keys); i++ {
						addkeys = append(addkeys, decodeHexString(t, update.Keys[i]))
						addvalues = append(addvalues, decodeHexString(t, update.Values[i]))
					}

					current, err := trie.GetCurrentTrie(txn)
					if err != nil {
						return err
					}

					stateRoot, err := current.Update(txn, addkeys, addvalues)
					if err != nil {
						return err
					}

					if err := setRootForHeight(txn, uint32(height), stateRoot); err != nil {
						return err
					}
					if err := SetCurrentStateRoot(txn, stateRoot); err != nil {
						return err
					}
					_, err = current.Commit(txn, uint32(height))
					if err != nil {
						return err
					}

					height++
				}

				return nil
			})
			require.Nil(t, err)

			err = db.View(func(txn *badger.Txn) error {
				lastUpdate := testData.Updates[len(testData.Updates)-1]
				lastKey := lastUpdate.Keys[len(lastUpdate.Keys)-1]
				utxoID, err := utils.DecodeHexString(lastKey)
				require.Nil(t, err)
				_, missing, err := trie.Get(txn, [][]byte{utxoID})
				if len(missing) != 0 {
					return fmt.Errorf("not found: %s", lastKey)
				}
				return err
			})
			assert.Nil(t, err)
		}()
	}
}

func decodeHexString(t *testing.T, value string) []byte {
	t.Helper()
	result, err := utils.DecodeHexString(value)
	require.Nil(t, err)
	return utils.CopySlice(result)
}
