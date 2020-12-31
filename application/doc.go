package application

/*
TX RULES
  1.  The sum of value in vout MUST be equal to the sum of value in vin with
      deposit discounts applied.
  2.  The txHash MUST be correct for the transaction
  3.  All signatures must be valid
  4.  No conflicting transaction hashes are allowed
  5.  No conflicting UTXO IDs are allowed
  6.  All referenced UTXOs must be valid and not be spent OR reference a
      deposit from Ethereum.
  9.  DataStore deposit must GTE to the correct amount based on data size
      and be mod zero datasize
  10. If the transaction is being mined, a datastore issuedAt must be equal to
      the current epoch
  11. The datastore index MUST be 32 bytes
  12. The chainID MUST be correct
  13. All transactions MUST have more than one input and more than one output.
*/

/*
RULES:
  THE SUM OF THE INPUTS MUST BE EQUAL TO THE SUM OF THE OUTPUTS.
  ALL INPUTS MUST REFERENCE ANOTHER UTXO/DTXO
  ALL OUTPUTS MUST NOT COLLIDE WITH ANY EXISTING OUTPUT IN HASH KEY

  THE REWARD FOR CLEANING UP A DTXO IS ONE EPOCH, IT MAY BE CLAIMED
  AT ANY TIME IN THE LAST EPOCH BY ANYONE

sync procedure - sync the latest snapshot trie
use the leaves of each subtree to ask for more
leaves

NOTE: this implies mined transactions must be indexed by leaf hashes
      same with DTXO data

  store transactions at prefixTx|hash(transaction)
  store UTXO DTXO as prefixUTXO|hash(transaction)|hash(transaction|index)
  store DTXO index as prefixUTXODataIndex|hash(transaction)|hash(transaction|index)

  store pending tx as a blob with data included
      use ref counting to prevent double inclusion

  for pending transactions maintain a set of index keys stored as follows:
      FOR DTXO containing tx:
          Note at this time miner rewards are done through DTXO objects
          as such all tx should contain a dtxo if they have a reward.
          This logic will cause some valid tx to potentially be dropped across
          epoch boundaries.
          Time and epoch based sorting allows stale data to be dropped
          back referencing allows data to be cleaned up when a tx is
          consumed. Time references also allow txs to be consumed in a
          FIFO manner of preference for initial deployment.
            key:
                prefixPendingTxIndex|epoch issued|ts received|hash(tx)
            value:
                Tx
            key:
                prefixPendingTxBackRef|hash(tx)
            value:
                prefixPendingTxIndex|epoch issued|ts received|hash(tx)


  Mined transaction indexing:


        key:
            prefixMinedTxIndex|epoch mined|hash(tx)
        value:
            Tx
        key:
            prefixMinedTxBackRef|hash(tx)
        value:
            prefixMinedTxIndex|epoch issued|hash(tx)

  UTXO indexing

        key:
            prefixUTXO|epoch expires|deposit value|hash(utxo)

            *Note: deposit value is used here to do greedy garbage
            collection. This allows the most space heavy objects to
            be consumed in a preferential manner. Thus rational
            actors will clean up the system in the most efficient
            manner possible in order to maximize rewards.
        value:
            utxo/dtxo (note data not stored here)

        key:
            prefixUTXO|epoch expires|epoch mined|hash(utxo)
        value:
            prefixUTXO|epoch expires|epoch mined|hash(utxo)
        key:
            prefixUTXOTxRef|



        must be indexed by hash
        must be indexed by blocknumber and index
        data must be indexed by expiration epoch



sync strategy:

                0
          /         \
        1            8
      /\           /   \
    2   5         9    10
  /\    /\       /\     /\
3   4  6  7   11  12  13  14
*/
