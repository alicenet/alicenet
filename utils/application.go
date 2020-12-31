package utils

import "encoding"

// GetObjSize returns the size of a marshalled capnproto object.
func GetObjSize(obj encoding.BinaryMarshaler) (int, error) {
	d, err := obj.MarshalBinary()
	if err != nil {
		return 0, err
	}
	return len(d), nil
}
