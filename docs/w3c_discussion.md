# Overview of W3C Decentralized Identifiers and Verifiable Credentials

W3C has started the process of standardizing methods
for identifiers and credentials.
Identifiers and credentials are usually issued by centralized organizations,
but the distributed nature of the Internet
has made it desirable to have identifiers and credentials **decentralized**.
Furthermore, it would be beneficial for these to be
**cryptographically verifiable**.

We give an overview of the W3C Standards for Decentralized Identifiers
and Verifiable Credentials.
These standards present a broad architecture and allow for significant leeway in their implementation;
because of this,
we discuss the specifics used by some projects on AliceNet.

We note that while the decentralized identifiers and verifiable credential
specifications are designed independently,
it is understood that they may be used together.

## Decentralized Identifiers (DIDs)

### Overview of DID Specification

Many identifiers come from centralized entities;
this includes driver's licenses issued by governments.
[W3C's Decentralized Identifier (DID) specification](https://w3c.github.io/did-core/)
is meant to facilitate the existence of identifiers
which are separate from any issuing authority.
A DID document contains information related to the identifier.
This is meant to ensure the controller of the identifier
(the DID Controller)
determines how the identifier is shared and what is revealed.
The DID references an entity (the DID Subject),
and the DID may be shared as determined by the DID Controller.
Security is given by cryptographic proof (usually via a digital signature).

The DID specification requires an identifier
of the form `did:<method-name>:<method-specific-id>`;
different methods will have different requirements.

### Important Features for AliceNet

AliceNet will use the Valence method for DIDs;
they have the form `did:valence:<string>`,
where `<string>` is a valid `base64url` encoding of 32 bytes
prepended with `u` (and removing the trailing `==`).
This will lead to a total of 44 `base64url` characters:
the prepended `u` followed by 43 characters encoding the bytes.

While the W3C standard makes no restriction on the relationship between
the DID Subject and DID Controller,
in the Valence implementation it is expected that
**the DID Subject and the DID Controller will be the same**;
that is, the DID document will be self-signed
(like self-signed x509 certificates).
The verification method **must** include an ECDSA public key using
the `secp256k1` elliptic curve or an `ed25519` public key.
Additionally, the DID Identifier (DID ID)
will be derived from the public key
(specifically, the `SHA3-256` hash of the public key);
this is similar to how Ethereum addresses are derived from
Ethereum public keys.

In the case where entities will be issuing
[Verifiable Credentials](#verifiable-credentials-and-presentations-vcs-and-vps),
the associated DID will be stored within a smart contract on AliceNet;
this will serve as AliceNet's Public Key Infrastructure
(discussed [here](#alicenet-public-key-infrastructure)).
Furthermore, a `bls12-381` public key **must** be included
within the Issuer's DID document
(in order to validate the Verifiable Credentials).

## Verifiable Credentials and Presentations (VCs and VPs)

### Overview of VC Specification

There are many types credentials, and many of them are physical credentials.
This makes them problematic when interacting over the Internet
or any other decentralized setting,
as it is not clear **why** a credential should be trusted.
The [W3C Verifiable Credential specification](https://w3c.github.io/vc-data-model/)
seeks to make this easier;
the validity of a VC is based on validating the digital signature
or cryptographic proof coupled with trusting the Issuer.

After the subject Bob has the verifiable credential (VC),
he may present a verifiable presentation (VP)
when asked;
this discloses some information from the VC
along with a cryptographic proof.
If implementation method uses **selective-disclosure**,
then it is possible that only the minimal information will need to be disclosed
to the recipient.
As an example, if Bob has a VC of a driver's license,
it would be possible to reveal his age without revealing his home address.

There is no restriction on who can issue a credential.
After a credential has been issued,
the **value** of the credential depends upon which entity
acted as the issuing authority for the credential and the associated due diligence
which went into issuing it.

### Overview of VP Specification

The Verifiable Credential is expected to contain
private information.
Ideally, only the minimum-required information from a credential
should be shared.
This is possible if the Verifiable Credential is issued
with the ability to perform selective-disclosure.

We note there is no specified protocol for the selected disclosure of claims
within a VC within the
[W3C Verifiable Credential specification](https://w3c.github.io/vc-data-model/);
with that said, all VCs which are used within AliceNet will use BBS signatures,
and VPs will use non-interactive zero-knowledge (NIZK) BBS proofs;
a discussion of BBS signatures may be found [here](./bbs_discussion.md).
These NIZK BBS proofs allow for VPs to selectively disclose VC attributes.

### Important Features for AliceNet

All VC Issuers will have a self-signed DID document stored on AliceNet
for public verification;
these DID documents contain public keys which are used
to verify VCs and VPs.
The fact that AliceNet stores these DID documents
means that AliceNet acts as its own public key infrastructure
(this is discussed more [here](#alicenet-public-key-infrastructure)).

One project on AliceNet seeks to provide 
**data lineage** (tracking when and how data is used);
in order to do this,
it is necessary to assign a consistent identifier to every VP.
One way to do this is by ensuring VPs are generated
in a deterministic manner
(and computing the identifier based on the resulting NIZK BBS proof).


## AliceNet Public Key Infrastructure

Public Key Cryptography relies on the secure distribution
of public keys.

The security of DID documents, Verifiable Credentials,
and Verifiable Presentations
rely on the ability to verify digital signatures and cryptographic proofs.
To do this, it is necessary to trust the public key used for verification.
This is usually implemented using some form of Public Key Infrastructure (PKI).

To that end, AliceNet will host the DID documents within a smart contract
which provide the public keys necessary to verify signatures and proofs
for applications built on AliceNet;
thus, AliceNet will provide its own PKI.
Valence will handle DID Document registration and revocation;
both registration and revocation will occur only after authentication.