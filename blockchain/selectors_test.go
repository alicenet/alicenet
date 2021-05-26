package blockchain_test

import (
	"testing"

	"github.com/MadBase/MadNet/blockchain"
)

func TestSelector(t *testing.T) {
	selector := blockchain.Selector("fsd")

	t.Logf("selector:%x", selector)
}

func TestSignature(t *testing.T) {

}

func TestConcurrency(t *testing.T) {

}
