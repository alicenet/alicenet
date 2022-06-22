using Go = import "/go.capnp";
@0xb99093b7d2518300;
$Go.package("capn");
$Go.import("github.com/alicenet/alicenet/application/capn");

const defaultDSPreImage :DSPreImage = (chainID = 0, index = 0x"00", issuedAt = 0, deposit = 0, rawData = 0x"00", owner = 0x"00", deposit1 = 0, deposit2 = 0, deposit3 = 0, deposit4 = 0, deposit5 = 0, deposit6 = 0, deposit7 = 0, fee0 = 0, fee1 = 0, fee2 = 0, fee3 = 0, fee4 = 0, fee5 = 0, fee6 = 0, fee7 = 0);
const defaultDSLinker :DSLinker = (txHash = 0x"00", dSPreImage = .defaultDSPreImage);
const defaultVSPreImage :VSPreImage = (chainID = 0, value = 0, owner = 0x"00", value1 = 0, value2 = 0, value3 = 0, value4 = 0, value5 = 0, value6 = 0, value7 = 0, fee0 = 0, fee1 = 0, fee2 = 0, fee3 = 0, fee4 = 0, fee5 = 0, fee6 = 0, fee7 = 0);
const defaultASPreImage :ASPreImage = (chainID = 0, value = 0, owner = 0x"00", issuedAt = 0, exp = 0, value1 = 0, value2 = 0, value3 = 0, value4 = 0, value5 = 0, value6 = 0, value7 = 0, fee0 = 0, fee1 = 0, fee2 = 0, fee3 = 0, fee4 = 0, fee5 = 0, fee6 = 0, fee7 = 0);
const defaultTXInPreImage :TXInPreImage = (chainID = 0, consumedTxIdx = 0, consumedTxHash = 0x"00");
const defaultTXInLinker :TXInLinker = (tXInPreImage = .defaultTXInPreImage, txHash = 0x"00");

struct DSPreImage {
    chainID @0 :UInt32 = 0;
    # The chainID of the chain this object was created on.

    index @1 :Data = 0x"00";
    # The index of this data reference.

    issuedAt @2 :UInt32 = 0;
    # The Epoch during which this object was created.

    rawData @4 :Data = 0x"00";
    # The raw data associated with this data store.

    tXOutIdx @5 :UInt32 = 0;
    # The index at which this element appears in the transaction output list.

    owner @6 :Data = 0x"00";
    # The hash of the public key of the owner of this object.

    deposit @3 :UInt32 = 0;
    deposit1 @7 :UInt32 = 0;
    deposit2 @8 :UInt32 = 0;
    deposit3 @9 :UInt32 = 0;
    deposit4 @10 :UInt32 = 0;
    deposit5 @11 :UInt32 = 0;
    deposit6 @12 :UInt32 = 0;
    deposit7 @13 :UInt32 = 0;
    # Deposit stores the value of the DataStore

    fee0 @14 :UInt32 = 0;
    fee1 @15 :UInt32 = 0;
    fee2 @16 :UInt32 = 0;
    fee3 @17 :UInt32 = 0;
    fee4 @18 :UInt32 = 0;
    fee5 @19 :UInt32 = 0;
    fee6 @20 :UInt32 = 0;
    fee7 @21 :UInt32 = 0;
    # Fee stores the associated fee for a DataStore
}

struct DSLinker {
    dSPreImage @0 :DSPreImage = .defaultDSPreImage;
    # The structure containing particular information for this object.

    txHash @1 :Data = 0x"00";
    # The hash of the transaction that created this object.
}

struct DataStore {
    dSLinker @0 :DSLinker = .defaultDSLinker;
    # Linking from object to txHash.

    signature @1 :Data = 0x"00";
    # Signature of the DSLinker
}

################################################################################

struct VSPreImage {
    chainID @0 :UInt32 = 0;
    # The chainID of this object.

    tXOutIdx @2 :UInt32 = 0;
    # The index at which this element appears in the transaction output list.

    owner @3 :Data = 0x"00";
    # The hash of the public key of the owner of this object.

    value @1 :UInt32 = 0;
    value1 @4 :UInt32 = 0;
    value2 @5 :UInt32 = 0;
    value3 @6 :UInt32 = 0;
    value4 @7 :UInt32 = 0;
    value5 @8 :UInt32 = 0;
    value6 @9 :UInt32 = 0;
    value7 @10 :UInt32 = 0;
    # Value stores the value

    fee0 @11 :UInt32 = 0;
    fee1 @12 :UInt32 = 0;
    fee2 @13 :UInt32 = 0;
    fee3 @14 :UInt32 = 0;
    fee4 @15 :UInt32 = 0;
    fee5 @16 :UInt32 = 0;
    fee6 @17 :UInt32 = 0;
    fee7 @18 :UInt32 = 0;
    # Fee stores the associated fee for a ValueStore
}

struct ValueStore {
    vSPreImage @0 :VSPreImage = .defaultVSPreImage;
    # The structure containing particular information for this object.

    txHash @1 :Data = 0x"00";
    # The hash of the transaction that created this object.
}

################################################################################

struct ASPreImage {
    chainID @0 :UInt32 = 0;
    # The chainID of this object.

    tXOutIdx @2 :UInt32 = 0;
    # The index at which this element appears in the transaction output list.

    owner @3 :Data = 0x"00";
    # <sva><curve><hashlock><initial owner pubk hash><partner pubk hash>
    # The hash of the public key of the original owner of this object.

    issuedAt @4 :UInt32 = 0;
    # The Epoch during which this object was created.

    exp @5 :UInt32 = 0;
    # The Epoch during which this object will fall back to the original owner
    # if it is not claimed by the partner before this point. For safety this
    # should be at least three epochs after issuedAt.

    value @1 :UInt32 = 0;
    value1 @6 :UInt32 = 0;
    value2 @7 :UInt32 = 0;
    value3 @8 :UInt32 = 0;
    value4 @9 :UInt32 = 0;
    value5 @10 :UInt32 = 0;
    value6 @11 :UInt32 = 0;
    value7 @12 :UInt32 = 0;
    # Value stores the value which is to be swapped

    fee0 @13 :UInt32 = 0;
    fee1 @14 :UInt32 = 0;
    fee2 @15 :UInt32 = 0;
    fee3 @16 :UInt32 = 0;
    fee4 @17 :UInt32 = 0;
    fee5 @18 :UInt32 = 0;
    fee6 @19 :UInt32 = 0;
    fee7 @20 :UInt32 = 0;
    # Fee stores the associated fee for an AtomicSwap
}

struct AtomicSwap {
    aSPreImage @0 :ASPreImage = .defaultASPreImage;
    # The structure containing particular information for this object.

    txHash @1 :Data = 0x"00";
    # The hash of the transaction that created this object.
}

################################################################################

struct TXInPreImage {
    chainID @0 :UInt32 = 0;
    # Chain id on which this object was created.

    consumedTxIdx @1 :UInt32 = 0;
    # Index at which the consumed object was created in the tx named by
    # consumedTxHash or the max value of uint32 to signal a deposit from
    # Ethereum.

    consumedTxHash @2 :Data = 0x"00";
    # The hash of the transaction that created the object to be consumed
    # or the nonce of the deposit if input is a deposit from Ethereum bc.
}

struct TXInLinker {
    tXInPreImage @0 :TXInPreImage = .defaultTXInPreImage;
    # The structure containing particular information for this object.

    txHash @1 :Data = 0x"00";
    # The hash of the transaction that is consuming this object.
}

struct TXIn {
    tXInLinker @0 :TXInLinker = .defaultTXInLinker;
    # Linking from object to txHash.

    signature @1 :Data = 0x"00";
    # Signature of linker.
}

################################################################################

struct TXOut {
    union {
        dataStore  @0 :DataStore;
        # The output if it is a datastore

        valueStore  @1 :ValueStore;
        # The output if it is a valuestore

        atomicSwap @2 :AtomicSwap;
        # The output if it is an atomicswap
    }
}

################################################################################

struct Tx {
    vin @0 :List(TXIn) = [];
    # Transaction input vector.

    vout @1 :List(TXOut) = [];
    # Transaction output vector.

    fee0 @2 :UInt32 = 0;
    fee1 @3 :UInt32 = 0;
    fee2 @4 :UInt32 = 0;
    fee3 @5 :UInt32 = 0;
    fee4 @6 :UInt32 = 0;
    fee5 @7 :UInt32 = 0;
    fee6 @8 :UInt32 = 0;
    fee7 @9 :UInt32 = 0;
    # Fee stores the fee
}

################################################################################
