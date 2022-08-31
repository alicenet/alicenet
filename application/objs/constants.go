package objs

// SVA is the defined type utilized for designation of a Signature Verification
// Algorithm.
type SVA uint8

const (
	// ValueStoreSVA is the constant which specifies the
	// Signature Verification Algorithm used for ValueStore objects.
	ValueStoreSVA SVA = 1

	// DataStoreSVA is the constant which specifies the
	// Signature Verification Algorithm used for DataStore objects.
	DataStoreSVA = 3
)
