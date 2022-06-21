package blockchain

import (
	"sync"

	"github.com/alicenet/alicenet/blockchain/interfaces"
	"github.com/ethereum/go-ethereum/crypto"
)

type SelectorMapDetail struct {
	sync.RWMutex
	signatures map[interfaces.FuncSelector]string
	selectors  map[string]interfaces.FuncSelector
}

func NewSelectorMap() *SelectorMapDetail {
	return &SelectorMapDetail{
		signatures: make(map[interfaces.FuncSelector]string, 20),
		selectors:  make(map[string]interfaces.FuncSelector, 20)}
}

func NewKnownSelectors() interfaces.SelectorMap {
	sm := &SelectorMapDetail{
		signatures: make(map[interfaces.FuncSelector]string, 20),
		selectors:  make(map[string]interfaces.FuncSelector, 20)}

	// This let's us recognize contract deploy transactions
	byteCodePrefix := interfaces.FuncSelector{0x60, 0x80, 0x60, 0x40}

	sm.selectors["_deploy_"] = byteCodePrefix
	sm.signatures[byteCodePrefix] = "_deploy_"

	// This is for simple ETH transfers
	transferSelector := interfaces.FuncSelector{0x0, 0x0, 0x0, 0x0}

	sm.selectors["_transfer_"] = transferSelector
	sm.signatures[transferSelector] = "_transfer_"

	// Facet admin
	sm.Selector("addFacet(bytes4,address)")
	sm.Selector("removeFacet(bytes4)")
	sm.Selector("replaceFacet(byte4,address)")

	// Registry
	sm.Selector("register(string,address)")
	sm.Selector("remove(string)")

	// Tokens
	sm.Selector("approve(address,uint256)")
	sm.Selector("transfer(address,uint256)")

	// Staking related
	sm.Selector("initializeStaking(address)")
	sm.Selector("balanceReward()")
	sm.Selector("balanceRewardFor(address)")
	sm.Selector("balanceStake()")
	sm.Selector("balanceStakeFor(address)")
	sm.Selector("balanceUnlocked()")
	sm.Selector("balanceUnlockedFor(address)")
	sm.Selector("balanceUnlockedReward()")
	sm.Selector("balanceUnlockedRewardFor(address)")
	sm.Selector("lockStake(uint256)")
	sm.Selector("majorFine(address)")
	sm.Selector("majorStakeFine()")
	sm.Selector("minimumStake()")
	sm.Selector("minorFine(address)")
	sm.Selector("minorStakeFine()")
	sm.Selector("requestUnlockStake()")
	sm.Selector("rewardAmount()")
	sm.Selector("rewardBonus()")
	sm.Selector("setMajorStakeFine(uint256)")
	sm.Selector("setMinimumStake(uint256)")
	sm.Selector("setMinorStakeFine(uint256)")
	sm.Selector("setRewardAmount(uint256)")
	sm.Selector("setRewardBonus(uint256)")
	sm.Selector("unlockStake(uint256)")

	// Snapshot maintenance
	sm.Selector("initializeSnapshots(address)")

	sm.Selector("snapshot(bytes,bytes)")
	sm.Selector("setMinEthSnapshotSize(uint256)")
	sm.Selector("minEthSnapshotSize()")
	sm.Selector("setMinMadSnapshotSize(uint256)")
	sm.Selector("minMadSnapshotSize()")
	sm.Selector("setEpoch(uint256)")
	sm.Selector("epoch()")
	sm.Selector("getChainIdFromSnapshot(uint256)")
	sm.Selector("getRawBlockClaimsSnapshot(uint256)")
	sm.Selector("getRawSignatureSnapshot(uint256)")
	sm.Selector("getHeightFromSnapshot(uint256)")
	sm.Selector("getMadHeightFromSnapshot(uint256)")

	// Validator maintenance
	sm.Selector("initializeParticipants(address)")
	sm.Selector("addValidator(address,uint256[2])")
	sm.Selector("removeValidator(address,uint256[2])")
	sm.Selector("isValidator(address)")
	sm.Selector("getValidatorPublicKey(address)")
	sm.Selector("confirmValidators()")
	sm.Selector("validatorMaxCount()")
	sm.Selector("validatorCount()")
	sm.Selector("setValidatorMaxCount(uint8)")

	// Accusation management
	sm.Selector("AccuseMultipleProposal(bytes,bytes,bytes,bytes)")

	// EthDKG
	sm.Selector("initializeEthDKG(address)")
	sm.Selector("initializeState()")
	sm.Selector("updatePhaseLength(uint256)")

	sm.Selector("submit_master_public_key(uint256[4])")

	sm.Selector("getPhaseLength()")
	sm.Selector("initialMessage()")
	sm.Selector("initialSignatures(address,uint256)")
	sm.Selector("T_REGISTRATION_END()")
	sm.Selector("T_SHARE_DISTRIBUTION_END()")
	sm.Selector("T_DISPUTE_END()")
	sm.Selector("T_KEY_SHARE_SUBMISSION_END()")
	sm.Selector("T_MPK_SUBMISSION_END()")
	sm.Selector("T_GPKJ_SUBMISSION_END()")
	sm.Selector("T_GPKJDISPUTE_END()")
	sm.Selector("T_DKG_COMPLETE()")
	sm.Selector("publicKeys(address,uint256)")
	sm.Selector("isMalicious(address)")
	sm.Selector("shareDistributionHashes(address)")
	sm.Selector("keyShares(address,uint256)")
	sm.Selector("commitments_1st_coefficient(address,uint256)")
	sm.Selector("gpkj_submissions(address,uint256)")
	sm.Selector("master_public_key(uint256)")
	sm.Selector("numberOfRegistrations()")
	sm.Selector("addresses(uint256)")

	sm.Selector("Group_Accusation_GPKj(uint256[],uint256[],uint256[])")
	sm.Selector("Group_Accusation_GPKj_Comp(uint256[][],uint256[2][][],uint256,address)")

	sm.Selector("submit_dispute(address,uint256,uint256,uint256[],uint256[2][],uint256[2],uint256[2])")

	sm.Selector("submit_key_share(address,uint256[2],uint256[2],uint256[4])")
	sm.Selector("register(uint256[2])")
	sm.Selector("Submit_GPKj(uint256[4],uint256[2])")
	sm.Selector("distribute_shares(uint256[],uint256[2][])")

	sm.Selector("Successful_Completion()")

	return sm
}

func (selectorMap *SelectorMapDetail) Selector(signature string) interfaces.FuncSelector {

	// First check if we already have it
	selectorMap.RLock()
	selector, present := selectorMap.selectors[signature]
	selectorMap.RUnlock()
	if present {
		return selector
	}

	// Calculate and store value
	selector = CalculateSelector(signature)

	selectorMap.Lock()
	selectorMap.signatures[selector] = signature
	selectorMap.Unlock()

	return selector
}

func (selectorMap *SelectorMapDetail) Signature(selector interfaces.FuncSelector) string {
	selectorMap.RLock()
	defer selectorMap.RUnlock()

	return selectorMap.signatures[selector]
}

// CalculateSelector calculates the hash of the supplied function signature
func CalculateSelector(signature string) interfaces.FuncSelector {
	var selector [4]byte

	selectorSlice := crypto.Keccak256([]byte(signature))[:4]
	selector[0] = selectorSlice[0]
	selector[1] = selectorSlice[1]
	selector[2] = selectorSlice[2]
	selector[3] = selectorSlice[3]

	return selector
}

func ExtractSelector(data []byte) interfaces.FuncSelector {
	var selector [4]byte

	if len(data) >= 4 {
		for idx := 0; idx < 4; idx++ {
			selector[idx] = data[idx]
		}
	}

	return selector
}
