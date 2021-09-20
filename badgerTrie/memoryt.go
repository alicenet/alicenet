package trie

type MemoryTrie struct {
	smt *SMT
}

func NewMemoryTrie() *MemoryTrie {
	return &MemoryTrie{NewSMT(nil, Hasher, func() []byte { return []byte("!!") })}
}

func (mt *MemoryTrie) Update(keys, values [][]byte) ([]byte, error) {
	roothash, err := mt.smt.Update(nil, keys, values)
	if err != nil {
		return nil, err
	}
	mt.smt.Discard()
	return roothash, nil
}
