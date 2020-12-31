package objs

// Vin is a vector of TxIn objects
type Vin []*TXIn

// UTXOID returns the list of UTXOIDs from each TXIn in Vin
func (vin Vin) UTXOID() ([][]byte, error) {
	ids := [][]byte{}
	for i := 0; i < len(vin); i++ {
		id, err := vin[i].UTXOID()
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// PreHash returns the list of PreHashs from each TXIn in Vin
func (vin Vin) PreHash() ([][]byte, error) {
	ids := [][]byte{}
	for i := 0; i < len(vin); i++ {
		id, err := vin[i].PreHash()
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// IsDeposit returns a list of bools specifying if each TXTn in Vin
// is a deposit
func (vin Vin) IsDeposit() []bool {
	ids := []bool{}
	for i := 0; i < len(vin); i++ {
		id := vin[i].IsDeposit()
		ids = append(ids, id)
	}
	return ids
}
