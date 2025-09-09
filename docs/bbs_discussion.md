# Overview of BBS Signatures used in Verifiable Credentials and Presentations

BBS signatures (a type of digital signature)
make it possible to sign multiple messages
at the same time and then cryptographically prove the validity of a subset
of those messages.
This is used within Verifiable Credentials to selectively reveal information
(with results returned in a Verifiable Presentation)
and be certain that it is correct (with cryptographic proof).
We give an overview of BBS signatures (used within Verifiable Credentials)
and associated non-interactive zero-knowledge (NIZK) proofs
(used within Verifiable Presentations).

The security proof of the BBS signature scheme and associated
NIZK proof may be found in [this paper](https://eprint.iacr.org/2023/275).
The original BBS signature publication may be found
[here](https://eprint.iacr.org/2004/174).

## Review of BLS Signatures

In order to ease into the **BBS** signature scheme,
we will first discuss the **BLS** signature scheme.
Both schemes use a **bilinear pairing**.

BLS signatures were originally presented
[here](https://link.springer.com/content/pdf/10.1007/s00145-004-0314-9.pdf).

### Bilinear Pairing

A **bilinear pairing** is a mapping
$e:\mathbb{G}_{1}\times\mathbb{G}_{2}\to\mathbb{G}_{T}$
such that

$$
e(h_{1}^{\alpha}, h_{2}^{\beta}) = [e(h_{1}, h_{2})]^{\alpha\beta}
$$

for all $h_{i}\in\mathbb{G}_{i}$ and $\alpha,\beta\in\mathbb{F}_{p}$.
Here, $\mathbb{G}_{1}$, $\mathbb{G}_{2}$, and $\mathbb{G}_{T}$
are cyclic groups of order $p$,
and $\mathbb{F}_{p} = \{0, 1, \dots, p-1\}$ is the field of order $p$.

We let $g_{1}\in\mathbb{G}_{1}$ and $g_{2}\in\mathbb{G}_{2}$
be standard generators.

### BLS Signatures

To sign a message,
we need a hash function $H:\{0,1\}^{*}\to\mathbb{G}_{1}$.
Given a private key $x\in\mathbb{F}_{p}$,
the associated public key is $X = g_{2}^{x}$.

To sign a message $m\in\{0,1\}^{*}$,
we set

$$
\sigma = [H(m)]^{x}.
$$

To verify a message $m$ was signed by the private key
associated with the public key $X$,
we confirm

$$
e(\sigma, g_{2}) = e(H(m), X).
$$


## BBS Signature Scheme

We present the BBS signature scheme by first showing how to sign one message
before turning the general case of signing multiple messages.
One important property of the signature scheme is that the signature size is
**independent** of the number of messages being signed.

Verifiable Credentials are signed using a BBS signature.

### One Message

#### Signing

We let $h_{1}\in\mathbb{G}_{1}$ be an independent generator in $\mathbb{G}_{1}$;
that is, the value $\alpha_{1}\in\mathbb{F}_{p}$ such that
$h_{1} = g_{1}^{\alpha_{1}}$ is unknown.
We let $x\in\mathbb{F}_{p}$ be the private key
with associated public key $X = g_{2}^{x}$
and $m_{1}\in\mathbb{F}_{p}$ the message to be signed.

1.  Set $C = g_{1}h_{1}^{m_{1}}$
2.  Choose random $e\in\mathbb{F}_{p}$.
3.  Set $A = C^{1/(x + e)}$.
4.  The signature is $\sigma = (A, e)$.

We are assuming the message to be signed is an element
of a finite field $\mathbb{F}_{p}$;
this assumption is not problematic,
for if we want to sign $M\in\{0,1\}^{*}$ it is easy to construct a hash function
$H_{p}:\{0,1\}^{*}\to\mathbb{F}_{p}$
(see [RFC 9380](https://www.rfc-editor.org/rfc/rfc9380.html))
and then sign $m = H_{p}(M)$
(standard cryptographic assumptions about hash functions will ensure this is secure).

#### Verifying

We verify a signature $\sigma = (A, e)$ for a message $m_{1}$
and public key $X$ with the following algorithm:

1.  Set $C = g_{1}h_{1}^{m_{1}}$
2.  Verify

$$
e(A, g_{2}^{e}X) = e(C, g_{2}).
$$

### Signing Multiple Messages

#### Signing

We let $h_{1},\dots,h_{k}\in\mathbb{G}_{1}$ be $k$ independent generators
in $\mathbb{G}_{1}$.
We let $x\in\mathbb{F}_{p}$ be the private key
with associated public key $X = g_{2}^{x}$
and $m_{1},\dots,m_{k}\in\mathbb{F}_{p}$ be the messages to be signed.

1.  Set $C = g_{1}h_{1}^{m_{1}}\cdots h_{k}^{m_{k}}$
2.  Choose random $e\in\mathbb{F}_{p}$.
3.  Set $A = C^{1/(x + e)}$.
4.  The signature is $\sigma = (A, e)$.

#### Verifying

We verify a signature $\sigma = (A, e)$ for messages $m_{1}, \dots, m_{k}$
and public key $X$ with the following algorithm:

1.  Set $C = g_{1}h_{1}^{m_{1}}\cdots h_{k}^{m_{k}}$
2.  Verify

$$
e(A, g_{2}^{e}X) = e(C, g_{2}).
$$

### Constructing Independent Generators

We construct independent generators in $\mathbb{G}_{1}$
deterministically using a hash-to-curve algorithm based on
[RFC 9380](https://www.rfc-editor.org/rfc/rfc9380.html).
This makes it possible to store these values instead
of computing them on the fly for more efficient signing and verification.


## NIZK BBS Proof Scheme

A non-interactive zero-knowledge (NIZK) BBS proof
makes it possible to show that certain messages are contained
in a BBS signature.
This is done without revealing the original signature nor revealing
any of the other messages.
We recall that the proof is **non-interactive** because, once generated,
the Verifier does not need to interact with the Prover,
and **zero-knowledge** because the Verifier only learns
that the statement is true
(namely, the subset of messages were signed by the Prover).

We use is protocol from [this paper](https://eprint.iacr.org/2023/275).

Verifiable Presentations contain an NIZK BBS proof.

### NIZK BBS Proof Generation

Suppose we have a signature $\sigma = (A, e)$ for messages $m_{1}, \dots, m_{k}$
for the public key $X = g_{2}^{x}$.
For $J \subseteq \{1, \dots, k\}$,
we can provide proof that messages $m' = \{m_{j} \mid j\in J\}$
are part of $\sigma$.

We perform the following:

1.  Set $I = \{1, \dots, k\} \setminus J$
2.  Set $C = g_{1}h_{1}^{m_{1}}\cdots h_{k}^{m_{k}}$
3.  Choose random $r \leftarrow \mathbb{F}_{p}$
4.  Set $\overline{A} = A^{r}$
5.  Set $\overline{B} = [CA^{-e}]^{r}$
6.  Set $C_{J} = g_{1}\prod_{j\in J} h_{j}^{m_{j}}$
7.  Choose random $\alpha,\beta \leftarrow \mathbb{F}_{p}$
8.  Choose random $\delta_{i} \leftarrow \mathbb{F}_{p}$ for $i\in I$
9.  Set $U = C_{J}^{\alpha}\overline{A}^{\beta} \prod_{i\in I} h_{i}^{\delta_{i}}$
10. Set $c = \textsf{Hash}(m', \overline{A}, \overline{B}, U)$
11. Set $s = \alpha + rc$
11. Set $t = \beta - ec$
12. Set $u_{i} = \delta_{i} + rm_{i}c$ for $i\in I$
13. Set $\pi = (\overline{A}, \overline{B}, c, s, t, \{u_{i}\})$

The NIZK BBS proof is $\pi$.

### NIZK BBS Proof Verification

Given an NIZK BBS proof $\pi$ and public key $X$,
we verify the proof using the following protocol:

1.  Extract $(\overline{A}, \overline{B}, c, s, t, \{u_{i}\}) = \pi$
2.  Verify $e(\overline{A}, X) = e(\overline{B}, g_{2})$;
    if this fails, the proof is invalid.
3.  Set $I = \{1, \dots, k\} \setminus J$
4.  Set $C_{J} = g_{1}\prod_{j\in J} h_{j}^{m_{j}}$
5.  Set $c = \textsf{Hash}(m', \overline{A}, \overline{B}, U)$
6.  Verify $U = \overline{B}^{-c}C_{J}^{s}\overline{A}^{t}\prod_{i\in I}h_{i}^{u_{i}}$;
    if this fails, the proof is invalid.


## Implementation Notes

### Hash Function

The general hash function we use is `sha3-256`
from the [SHA3 standard](https://csrc.nist.gov/pubs/fips/202/final).

### Bilinear Pairing

We use the `bls12-381` bilinear pairing;
the BLS family of elliptic curves were originally defined
[here](https://eprint.iacr.org/2002/088).
We note that BLS **signatures**
(described [here](https://link.springer.com/article/10.1007/s00145-004-0314-9))
are different (although they use a bilinear pairing),
and that Ben **L**ynn is the only author on both papers.

### Randomness

Following [RFC 6979](https://www.rfc-editor.org/rfc/rfc6979.html),
we use a deterministic method to generate randomness
when generating BBS signatures and NIZK BBS proofs.
While not required by
[W3C Verifiable Credential and Presentation specification](https://w3c.github.io/vc-data-model/),
this determinism makes it possible to associate an identifier
with each VC and VP.