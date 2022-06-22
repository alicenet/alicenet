package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"

	"github.com/alicenet/alicenet/crypto"
	eth "github.com/ethereum/go-ethereum/crypto"
)

func main() {
	modePtr := flag.String("mode", "sm", "One of sm for sign message sh for sign hash, vm for validate message or vh for validate hash.")
	sigPtr := flag.String("s", "", "Signature.")
	mPtr := flag.String("m", "", "Message as string.")
	hshPtr := flag.String("h", "", "Hash as hex.")
	pPtr := flag.String("p", "", "Private key as hex.")
	flag.Parse()
	mode := *modePtr
	switch mode {
	case "vh":
		sigBytes, err := hex.DecodeString(*sigPtr)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		mBytes, err := hex.DecodeString(*hshPtr)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		pubk, err := eth.SigToPub(mBytes, sigBytes)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		pubkbytes := eth.FromECDSAPub(pubk)
		account := crypto.GetAccount(pubkbytes)
		fmt.Printf("Signature is good. Signed by account: %x\n", account)
	case "vm":
		sigBytes, err := hex.DecodeString(*sigPtr)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("# sig bytes: %v\n", len(sigBytes))
		mBytes := []byte(*mPtr)
		fmt.Printf("# msg bytes: %v\n", len(mBytes))
		hsh := crypto.Hasher(mBytes)
		pubk, err := eth.SigToPub(hsh, sigBytes)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		pubkbytes := eth.FromECDSAPub(pubk)
		account := crypto.GetAccount(pubkbytes)
		fmt.Printf("Signature is good. Signed by account: %x\n", account)
	case "sm":
		fmt.Println("Signing Message Mode")
		signer := crypto.Secp256k1Signer{}
		privkBytes, err := hex.DecodeString(*pPtr)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("# key bytes: %v\n", len(privkBytes))
		err = signer.SetPrivk(privkBytes)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		mBytes := []byte(*mPtr)
		fmt.Printf("The hash is %x\n", crypto.Hasher(mBytes))
		sig, err := signer.Sign(mBytes)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		pubkeybytes, err := signer.Pubkey()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Signature: %x\n", sig)
		fmt.Printf("Signed by account: %x\n", crypto.GetAccount(pubkeybytes))
	case "sh":
		fmt.Println("Signing hash Mode")
		privkBytes, err := hex.DecodeString(*pPtr)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("# key bytes: %v\n", len(privkBytes))
		mBytes, err := hex.DecodeString(*hshPtr)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		ecprivk, err := eth.ToECDSA(privkBytes)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		sig, err := eth.Sign(mBytes, ecprivk)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Signature: %x\n", sig)
		fmt.Printf("Signed by account: %x\n", crypto.GetAccount(eth.FromECDSAPub(&ecprivk.PublicKey)))
	default:
		fmt.Printf("Unknown mode: %v\n", mode)
		os.Exit(1)
	}
}
