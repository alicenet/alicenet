package objs

// SVA is the defined type utilized for designation of a Signature Verification
// Algorithm
type SVA uint8

const (
	// ValueStoreSVA is the constant which specifies the
	// Signature Verification Algorithm used for ValueStore objects
	ValueStoreSVA SVA = iota + 1

	// HashedTimelockSVA is the constant which specifies the
	// Signature Verification Algorithm used for AtomicSwap objects
	HashedTimelockSVA

	// DataStoreSVA is the constant which specifies the
	// Signature Verification Algorithm used for DataStore objects
	DataStoreSVA
)

// SignerRole is the defined type utilized for designation for signers
// in AtomicSwap objects
type SignerRole uint8

const (
	// PrimarySignerRole is the constant which specifies the role of the
	// primary account owner in the AtomicSwap object
	PrimarySignerRole SignerRole = iota + 1

	// AlternateSignerRole is the constant which specifies the role of the
	// alternate account owner in the AtomicSwap object
	AlternateSignerRole
)
