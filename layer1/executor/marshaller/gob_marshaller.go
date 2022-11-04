package marshaller

import (
	"bytes"
	"encoding/gob"

	"github.com/alicenet/alicenet/layer1/executor/tasks"
)

// var _ encoding.BinaryMarshaler = &tasks.Task{}

// todo: ticket to investigate migrating old TypeRegistry to GOB like here
func GobMarshalBinary(task tasks.Task) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(&task)
	if err != nil {
		return nil, err
	}
	vv := buf.Bytes()
	return vv, nil
}

func GobUnmarshalBinary(data []byte) (tasks.Task, error) {
	buf := &bytes.Buffer{}
	buf.Write(data)
	dec := gob.NewDecoder(buf)
	var acc tasks.Task
	err := dec.Decode(&acc) // decode concrete implementation into an interface var without knowing which implementation it is (gob is awesome)
	if err != nil {
		return nil, err
	}
	return acc, nil
}
