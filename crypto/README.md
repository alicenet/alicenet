# Crypto

## About

This directory holds all code related to cryptography.
**Any and all changes should be handled with great care.**

The primary cryptographic hash function used throughout AliceNet
is `Keccak256` as in Ethereum.
Any reference should call the hash function defined here.

## `bn256/`

This directory holds everything related to the `bn256` elliptic curve
(sometimes called `alt_bn128`).
It defines an bilinear pairing
associated with an elliptic curve over a finite field
with a 256-bit prime.
Although originally designed to provide 128 bits of security,
recent developments in factoring algorithms have reduced this security
to approximately 100 bits.

This is the one of the bilinear pairings Ethereum uses,
and we used their code as the foundation.
A hash-to-curve algorithm was added to enable BLS signatures.
The bilinear pairing enables enable threshold group signatures
for consensus.
Additional code related to the Distributed Key Generation protocol
has been added as well.

## `secp256k1/`

This directory holds everything related to the `secp256k1`
elliptic curve.
This is the elliptic curve used by Ethereum and Bitcoin
for digital signatures.
The code essentially wraps the `secp256k1` code from Ethereum.