#! /usr/bin/python3

# TODO: REWRITE IN TS

import binascii

class Cdef(object):
    def __init__(self, name, saltString):
        self.name = name
        self.saltString = saltString

def make(dat):
    out = []
    for d in dat:
        out.append(Cdef(d[0],d[1]))
    return out

def build(cdefs):
    ocurl = "{"
    ccurl = "}"
    outf = f"""
// SPDX-License-Identifier: MIT-open-group
pragma solidity 0.8.11;

import "./DeterministicAddress.sol";

abstract contract ImmutableFactory is DeterministicAddress {ocurl}

    address private immutable _factory;

    constructor(address factory_) {ocurl}
        _factory = factory_;
    {ccurl}

    modifier onlyFactory() {ocurl}
        require(msg.sender == _factory, "onlyFactory");
        _;
    {ccurl}

    function _factoryAddress() internal view returns(address) {ocurl}
        return _factory;
    {ccurl}

{ccurl}
"""
    tmpl = (
        lambda name, salt: f"""

abstract contract immutable{name} is ImmutableFactory {ocurl}

    address private immutable _{name};

    constructor() {ocurl}
        _{name} = getMetamorphicContractAddress(0x{salt}, _factoryAddress());
    {ccurl}

    modifier only{name}() {ocurl}
        require(msg.sender == _{name}, "only{name}");
        _;
    {ccurl}

    function _{name}Address() internal view returns(address) {ocurl}
        return _{name};
    {ccurl}

    function _saltFor{name}() internal pure returns(bytes32) {ocurl}
        return 0x{salt};
    {ccurl}

{ccurl}

"""
    )
    for c in cdefs:
        name = c.name
        salt = c.saltString.encode().ljust(32, binascii.unhexlify(b"00"))

        salt = binascii.hexlify(salt).decode("utf-8")
        render = tmpl(name, salt)
        outf = "".join([outf, render])
    return outf

def run():
    clst = [
        ("ValidatorNFT", "ValidatorNFT"),
        ("MadToken", "MadToken"),
        ("StakeNFT", "StakeNFT"),
        ("MadByte", "MadByte"),
        ("Governance", "Governance"),
        ("ValidatorPool", "ValidatorPool"),
        ("ETHDKG", "ETHDKG"),
        ("ETHDKGAccusations", "ETHDKGAccusations"),
        ("Snapshots", "Snapshots"),
        ("ETHDKGPhases", "ETHDKGPhases"),
        ("StakeNFTLP", "StakeNFTLP"),
        ("Foundation", "Foundation")
    ]
    c=make(clst)
    raw = build(c)
    print(raw)

run()

"""
replace last line in gomod
use dummy mad token to mint stake nft positions
EX: tests/tokens/periphery/validatorPool/business-logic.ts
register validators line 88 of above
finally validatorPool.initializeEthDKG

run golang
bridge: new-shapshot-contract
madnet: changes
"""
