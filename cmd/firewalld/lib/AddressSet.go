package lib

type AddressSet map[string]bool

func NewAddresSet(addrs []string) AddressSet {
	ret := make(map[string]bool)
	for _, addr := range addrs {
		ret[addr] = true
	}
	return ret
}

func (s AddressSet) String() string {
	ret := "["
	for addr := range s {
		if len(ret) > 1 {
			ret += ", "
		}
		ret += addr
	}
	return ret + "]"
}

func (s AddressSet) Has(addr string) bool {
	return s[addr]
}

func (s AddressSet) Add(addr string) {
	s[addr] = true
}

func (s AddressSet) Delete(addr string) {
	delete(s, addr)
}

func (s AddressSet) Equal(ss AddressSet) bool {
	if len(s) != len(ss) {
		return false
	}
	for i := range s {
		if !ss[i] {
			return false
		}
	}
	return true
}
