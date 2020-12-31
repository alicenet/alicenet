package types

// ChainIdentifier uniquely represents the chain that a client is
// interested in. No two peers should have a different chain
// identifier and be connected.
type ChainIdentifier uint32
