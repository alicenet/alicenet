package accusation

import (
	"fmt"

	aobjs "github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/consensus/lstate"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/dgraph-io/badger/v2"
)

func detectNonExistentUTXOConsumption(rs *objs.RoundState, lrs *lstate.RoundStates, db *db.Database) (tasks.Task, bool) {

	// if rs.Proposal.PClaims.BClaims.TxCount <= 0 {
	// 	return nil, false
	// }

	// rs.
	// if rs.Proposal.PClaims.BClaims.Height+1 != rs.Proposal.PClaims.BClaims.Height {
	// 	return nil, false
	// }
	var err error

	for _, txHash := range rs.Proposal.TxHshLst {
		// get Tx by hash from DB
		err = db.View(func(txn *badger.Txn) error {
			txBin, err := db.GetTxCacheItem(txn, rs.Proposal.PClaims.BClaims.Height, txHash)
			if err != nil {
				return err
			}

			tx := &aobjs.Tx{}
			err = tx.UnmarshalBinary(txBin)
			if err != nil {
				return err
			}

			utxoIDs, err := tx.Vin.UTXOID()
			if err != nil {
				return err
			}

			for _, utxoID := range utxoIDs {
				// check utxo id against state tree
				// if exists: continue, all is fine
				// else: get proof of exclusion - BOOM

			}
			/*
				accusation ID: keccak(height + UTXOID + StateTree root + txHash)
			*/

			return nil
		})

		if err != nil {
			panic(fmt.Sprintf("error getting tx: %v", err))
		}

		// check consumed tx.UTXOs are in State Tree
	}

	// bclaims + sigGroup
	// pclaims + sig
	// txInPreImage
	// [3][]byte proofs

	return nil, false
}
