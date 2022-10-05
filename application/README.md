# Application

## About

This directory holds everything related to `application`.
In particular, this holds everything related to transactions.

`application` is meant to be independent of `consensus`,
as the focus is on the current state and allowable (valid)
state transitions.
In that way, there is no real notion of time within `application`.
Time is tracked within `consensus`,
which keeps track of time by reaching consensus on the current state.

## `deposit/`

Here we store information related to the Deposit Handler.
This tracks all deposits received as well as all deposits consumed.
Deposits are unique `UTXO`s because their validity
**must come from the outside**:
it is easy to check if a deposit has been double spent,
but **the validity of a deposit requires a trusted source**.
In this case, the trusted source is from an event emission on Ethereum.

## `indexer/`

This holds the various indexers used within `application`.

## `minedtx/`

This stores information related to mined transactions;
that is, transactions which have been included in previous blocks.

## `objs/`

This contains all definitions related to transactions and `UTXO` types.

## `pendingtx/`

This stores information to handling pending transactions;
that is, transactions which provide valid changes to state.

## `utxohandler/`

This stores everything related to handling `UTXO`s.
One important function defined on the `utxohandler` is `IsValid`,
which determines if a collection of transactions form
a valid state transition.