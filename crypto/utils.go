package crypto

// GetAccount returns the account, which corresponds to the rightmost 20 bytes
// of the hash of the public key.
func GetAccount(pubk []byte) []byte {
	return Hasher(pubk[1:])[12:]
}

// CalcThreshold returns the correct definition of threshold.
// With this definition, t+1 participants form a BFT subset.
// This ensures
//			t <= 2/3*n		and 	t + 1 > 2/3*n
//
// Note Well: This convension for threshold (using t <= 2/3*n < t+1)
//			  is *different* other parts in the codebase.
//			  Within consensus, there is the convension that threshold
//			  number of participants ensure validity.
//			  Thus, within consensus, threshold refers to t+1
//			  using this convension.
func CalcThreshold(n int) int {
	k := n / 3
	t := 2 * k
	if (n - 3*k) == 2 {
		t = t + 1
	}
	return t
}
