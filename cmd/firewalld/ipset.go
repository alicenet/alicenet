package firewalld

type IPset map[string]bool

func NewIPset(ips []string) IPset {
	ret := make(map[string]bool)
	for _, ip := range ips {
		ret[ip] = true
	}
	return ret
}

func (s IPset) Add(ip string) {
	s[ip] = true
}

func (s IPset) Delete(ip string) {
	delete(s, ip)
}

func (s IPset) MarshallString() string {
	const comma byte = ','

	ret := make([]byte, 0)
	for ip := range s {
		ret = append(ret, comma)

		b := []byte(ip)
		ret = append(ret, b...)
	}
	ret = ret[1:] // remove first comma
	return string(ret)
}
