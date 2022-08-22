package cloudflare

import (
	"math/big"
)

// Sign generates a signature based on the message, private key, and
// Hash-to-G1 function
func Sign(msg []byte, privK *big.Int, hashG1Func func(msg []byte) (*G1, error)) (*G1, error) {
	sig := &G1{}
	hash, err := hashG1Func(msg)
	if err != nil {
		return nil, err
	}
	sig.ScalarMult(hash, privK)
	return sig, nil
}

// Verify ensures that
//
//	e(sig, h2) == e(hashG1Func(msg), pubK).
//
// To do this, we use PairingCheck because this will be called by the
// Ethereum Virtual Machine. Thus, we check
//
//	e(sig, h2Neg) * e(hashG1Func(msg), pubK) == 1.
//
// Here, h2Neg is the negation (in additive notation) of h2, which
// is the standard generator for G2.
//
// If we return a dangerous hash point (Infinity (the identity element)
// or the G1 generator or its negation), we cause verification
// to fail because signatures on these points can easily be forged.
// Hashing to these points break the security assumptions because
// the discrete logarithm is known.
func Verify(msg []byte, sig *G1, pubK *G2, hashG1Func func(msg []byte) (*G1, error)) (bool, error) {
	hash, err := hashG1Func(msg)
	if err != nil {
		if err == ErrDangerousPoint {
			return false, nil
		}
		return false, err
	}
	h2Neg := &G2{}
	h2Neg.p = &twistPoint{}
	h2Neg.p.Set(twistGenNeg)
	val := PairingCheck([]*G1{sig, hash}, []*G2{h2Neg, pubK})
	return val, nil
}

// LagrangeInterpolationG1 will interpolate the G1 values in pointsG1 and output
// the interpolated value. This is used in AggregateSignatures to combine
// partial signatures into the final group signature.
// Lagrange interpolation is the same as though we were performing
// interpolation over the real numbers, although in this instance we are
// interpolating over a finite field.
//
// The end result is
//
//	val = Prod(pointsG1[j]^Rj, j).
//
// Here, Rj is a constant which depends only on indices. As this relates to
// AggregateSignatures, the Rj constants only depend upon the signers
// themselves.
func LagrangeInterpolationG1(pointsG1 []*G1, indices []int, threshold int) (*G1, error) {
	if len(pointsG1) != len(indices) {
		return nil, ErrLIArrayMismatch
	}

	val := &G1{}
	for i, idxJ := range indices {
		if i > threshold {
			// We only require threshold+1 values, so we ignore all of the
			// values beyond this.
			break
		}
		// The full Rj constant is
		//
		//		Rj = Prod(k/(k-j), k != j),
		//
		// where the product is over all of the indices.
		// Here, we use k and j as shorthand for idxK and idxJ.
		// To build Rj, we loop over all indices and multiply each
		// RjPartial [k/(k-j)] together.
		Rj := big.NewInt(1)
		for ell, idxK := range indices {
			if ell > threshold {
				// See previous if-statement comment.
				break
			}
			if idxJ == idxK {
				// need to move on to next participant
				continue
			}
			RjPartial := liRjPartialConst(idxK, idxJ)
			Rj.Mul(Rj, RjPartial)
			Rj.Mod(Rj, Order)
		}
		partialVal := &G1{}
		partialVal.Set(pointsG1[i])
		partialVal.ScalarMult(partialVal, Rj)
		val.Add(val, partialVal)
	}
	return val, nil
}

// liRjPartialConst makes the Rj constants as required in
// LagrangeInterpolationG1. This partial constant is
//
//	k/(k - j).
//
// All of these operations are performed in the finite field F_Order
// (the finite field of size Order, the size of the groups G1, G2, and GT).
// This is *not* floating-point arithmetic. We assume k != j.
func liRjPartialConst(k, j int) *big.Int {
	tmp1 := big.NewInt(int64(k))
	tmp2 := big.NewInt(int64(k - j))
	tmp2.ModInverse(tmp2, Order)
	tmp2.Mul(tmp1, tmp2)
	tmp2.Mod(tmp2, Order)
	return tmp2
}

// AggregateSignatures computes the group signature from slices of signatures
// and public keys. In particular, given signature sig_j, we call
// LagrangeInterpolationG1 in order to produce
//
//	grpsig = Prod(sig_j^Rj, j).
//
// Here, Rj is a constant which depends only on the validating set.
// If mpk is the master public key (the group signing key), then we have
// the following equality:
//
//	e(grpsig, h2) == e(hashG1Func(M), mpk).
func AggregateSignatures(sigs []*G1, indices []int, threshold int) (*G1, error) {
	if len(sigs) != len(indices) {
		return nil, ErrMismatchedSlices
	}
	if len(sigs) <= threshold {
		return nil, ErrBelowThreshold
	}
	grpsig, err := LagrangeInterpolationG1(sigs, indices, threshold)
	return grpsig, err
}

// MarshalSignature takes in the sig and pubK and outputs byte slice of
// pubK bytes followed by sig bytes
func MarshalSignature(sig *G1, pubK *G2) ([]byte, error) {
	ret := []byte{}
	ret = append(ret, pubK.Marshal()...)
	ret = append(ret, sig.Marshal()...)
	return ret, nil
}

// UnmarshalSignature takes the marshalled signature and
// converts it back to the sig and pubK
func UnmarshalSignature(marshalledSig []byte) (*G2, *G1, error) {
	if len(marshalledSig) != 6*numBytes {
		return nil, nil, ErrInvalidSignatureLength
	}
	sig := &G1{}
	sig.p = &curvePoint{}
	pubK := &G2{}
	pubK.p = &twistPoint{}
	tmp, err := pubK.Unmarshal(marshalledSig)
	if err != nil {
		return nil, nil, err
	}
	_, err = sig.Unmarshal(tmp)
	if err != nil {
		return nil, nil, err
	}
	return pubK, sig, nil
}

// SplitPubkeySig separates the signature and public key bytes
// and returns them individually
func SplitPubkeySig(sig []byte) ([]byte, []byte, error) {
	if len(sig) != 6*numBytes {
		return nil, nil, ErrInvalidSignatureLength
	}
	pubkBytes := sig[0 : 4*numBytes]
	sigBytes := sig[4*numBytes:]
	return pubkBytes, sigBytes, nil
}

// PubkeyFromSig returns the public key bytes from a marshalled signature
func PubkeyFromSig(sig []byte) ([]byte, error) {
	pubkb, _, err := SplitPubkeySig(sig)
	if err != nil {
		return nil, err
	}
	return pubkb, nil
}

// AggregateMarshalledSignatures takes marshalled signatures and forms a group
// signature assuming there are no errors. See AggregateSignatures for more
// information specifically about signature aggregation.
//
// listOfPubKsMarsh is the list of (marshalled) public keys in the correct
// order. This allows the specific order to be chosen outside of the
// cryptographic routines. marshalledSigs contains both the signature and the
// corresponding public key, which enables us to call makeIndicesArray in
// order to correctly perform AggregateSignatures.
func AggregateMarshalledSignatures(marshalledSigs [][]byte, listOfPubKsMarsh [][]byte, threshold int) (*G1, error) {
	if len(marshalledSigs) <= threshold {
		return nil, ErrBelowThreshold
	}
	var err error
	sigs := make([]*G1, len(marshalledSigs))
	pubKs := make([]*G2, len(marshalledSigs))
	for k := 0; k < len(marshalledSigs); k++ {
		pubKs[k], sigs[k], err = UnmarshalSignature(marshalledSigs[k])
		if err != nil {
			return nil, err
		}
	}
	orderedPubKs := make([]*G2, len(listOfPubKsMarsh))
	for k := 0; k < len(listOfPubKsMarsh); k++ {
		orderedPubKs[k] = &G2{}
		orderedPubKs[k].p = &twistPoint{}
		_, err = orderedPubKs[k].Unmarshal(listOfPubKsMarsh[k])
		if err != nil {
			return nil, err
		}
	}

	indices, err := makeIndicesArray(pubKs, orderedPubKs)
	if err != nil {
		return nil, err
	}
	grpsig, err := AggregateSignatures(sigs, indices, threshold)
	return grpsig, err
}

// makeIndicesArray takes the public keys in pubKs and finds their correct
// index from orderedPubKs. Note that we add 1 to each index,
// because the indexing starts at 1. This is required so that
// shared secrets are shared and not given away.
func makeIndicesArray(pubKs []*G2, orderedPubKs []*G2) ([]int, error) {
	indices := make([]int, len(pubKs))
	for k := 0; k < len(pubKs); k++ {
		foundInt := false
		curPubK := pubKs[k]
		for j := 0; j < len(orderedPubKs); j++ {
			curOrderedPubK := orderedPubKs[j]
			if curPubK.IsEqual(curOrderedPubK) {
				indices[k] = j + 1
				foundInt = true
				break
			}
		}
		if foundInt {
			continue
		} else {
			return nil, ErrMissingIndex
		}
	}
	return indices, nil
}
