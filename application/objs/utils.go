package objs

import (
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/utils"
)

// MakeUTXOID will create the UTXOID for a utxo given a transaction hash and
// index.
func MakeUTXOID(txHash []byte, idx uint32) []byte {
	if idx == constants.MaxUint32 {
		return utils.CopySlice(txHash)
	}
	idxBytes := utils.MarshalUint32(idx)
	msg := utils.CopySlice(txHash)
	msg = append(msg, idxBytes...)
	return crypto.Hasher(msg)
}

func extractSignature(owner []byte, curveSpec constants.CurveSpec) ([]byte, []byte, error) {
	switch curveSpec {
	case constants.CurveSecp256k1:
		if len(owner) < constants.CurveSecp256k1SigLen {
			return nil, nil, errorz.ErrInvalid{}.New("extractSignature; Invalid Secp256k1 signature")
		}
		return owner[0:constants.CurveSecp256k1SigLen], utils.CopySlice(owner[constants.CurveSecp256k1SigLen:]), nil
	case constants.CurveBN256Eth:
		if len(owner) < constants.CurveBN256EthSigLen {
			return nil, nil, errorz.ErrInvalid{}.New("extractSignature; Invalid BN256Eth signature")
		}
		return owner[0:constants.CurveBN256EthSigLen], utils.CopySlice(owner[constants.CurveBN256EthSigLen:]), nil
	default:
		return nil, nil, errorz.ErrInvalid{}.New("extractSignature; Invalid curveSpec")
	}
}

func validateSignatureLen(sig []byte, curveSpec constants.CurveSpec) error {
	switch curveSpec {
	case constants.CurveSecp256k1:
		if len(sig) == constants.CurveSecp256k1SigLen {
			return nil
		}
		return errorz.ErrInvalid{}.New("validateSignatureLength; Invalid secp256k1 sig len")
	case constants.CurveBN256Eth:
		if len(sig) == constants.CurveBN256EthSigLen {
			return nil
		}
		return errorz.ErrInvalid{}.New("validateSignatureLength; Invalid bn256 sig len")
	default:
		return errorz.ErrInvalid{}.New("validateSignatureLength; Invalid curveSpec")
	}
}

func extractCurveSpec(owner []byte) (constants.CurveSpec, []byte, error) {
	if len(owner) < 1 {
		return 0, nil, errorz.ErrInvalid{}.New("extractCurveSpec; extraction failed")
	}
	return constants.CurveSpec(owner[0]), utils.CopySlice(owner[1:]), nil
}

func extractSignerRole(owner []byte) (SignerRole, []byte, error) {
	if len(owner) < 1 {
		return 0, nil, errorz.ErrInvalid{}.New("extractSignerRole; extraction failed")
	}
	return SignerRole(owner[0]), utils.CopySlice(owner[1:]), nil
}

func extractAccount(owner []byte) ([]byte, []byte, error) {
	if len(owner) < constants.OwnerLen {
		return nil, nil, errorz.ErrInvalid{}.New("extractAccount; extraction failed")
	}
	return owner[0:constants.OwnerLen], utils.CopySlice(owner[constants.OwnerLen:]), nil
}

func extractSVA(owner []byte) (SVA, []byte, error) {
	if len(owner) < 1 {
		return 0, nil, errorz.ErrInvalid{}.New("extractSVA: extraction failed")
	}
	return SVA(owner[0]), utils.CopySlice(owner[1:]), nil
}

func extractHash(owner []byte) ([]byte, []byte, error) {
	if len(owner) < constants.HashLen {
		return nil, nil, errorz.ErrInvalid{}.New("extractHash: extraction failed")
	}
	return owner[0:constants.HashLen], utils.CopySlice(owner[constants.HashLen:]), nil
}

func extractZero(owner []byte) error {
	if len(owner) != 0 {
		return errorz.ErrInvalid{}.New("extractZero; bytes remaining at end")
	}
	return nil
}
