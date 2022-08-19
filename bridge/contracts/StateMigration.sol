// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/ISnapshots.sol";
import "contracts/interfaces/IStakingToken.sol";
import "contracts/interfaces/IUtilityToken.sol";
import "contracts/interfaces/IValidatorPool.sol";
import "contracts/interfaces/IERC20Transferable.sol";
import "contracts/interfaces/IStakingNFT.sol";
import "contracts/interfaces/IETHDKG.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/libraries/parsers/BClaimsParserLibrary.sol";
import "@openzeppelin/contracts/token/ERC721/IERC721.sol";

contract ExternalStore is ImmutableFactory {
    uint256[4] internal _tokenIDs;
    uint256 internal _counter;

    constructor(address factory_) ImmutableFactory(factory_) {}

    function storeTokenIds(uint256[4] memory tokenIDs) public onlyFactory {
        _tokenIDs = tokenIDs;
    }

    function incrementCounter() public onlyFactory {
        _counter++;
    }

    function getTokenIds() public view returns (uint256[4] memory) {
        return _tokenIDs;
    }

    function getCounter() public view returns (uint256) {
        return _counter;
    }
}

contract StateMigration is
    ImmutableFactory,
    ImmutableSnapshots,
    ImmutableETHDKG,
    ImmutableAToken,
    ImmutableATokenMinter,
    ImmutableBToken,
    ImmutablePublicStaking,
    ImmutableValidatorPool
{
    uint256 public constant EPOCH_LENGTH = 1024;
    ExternalStore internal immutable _externalStore;

    constructor(address factory_)
        ImmutableFactory(factory_)
        ImmutableSnapshots()
        ImmutableETHDKG()
        ImmutableAToken()
        ImmutableBToken()
        ImmutableATokenMinter()
        ImmutablePublicStaking()
        ImmutableValidatorPool()
    {
        _externalStore = new ExternalStore(factory_);
    }

    function doMigrationStep() public {
        if (_externalStore.getCounter() == 0) {
            stakeValidators();
            _externalStore.incrementCounter();
            return;
        }
        if (_externalStore.getCounter() == 1) {
            migrateSnapshotsAndValidators();
            _externalStore.incrementCounter();
            return;
        }
    }

    function stakeValidators() public {
        // Setting staking amount
        IValidatorPool(_validatorPoolAddress()).setStakeAmount(1);
        // Minting 4 aTokensWei to stake the validators
        IStakingTokenMinter(_aTokenMinterAddress()).mint(_factoryAddress(), 4);
        IERC20Transferable(_aTokenAddress()).approve(_publicStakingAddress(), 4);
        uint256[4] memory tokenIDs;
        for (uint256 i; i < 4; i++) {
            // minting publicStaking position for the factory
            tokenIDs[i] = IStakingNFT(_publicStakingAddress()).mint(1);
            IERC721(_publicStakingAddress()).approve(_validatorPoolAddress(), tokenIDs[i]);
        }
        _externalStore.storeTokenIds(tokenIDs);
    }

    function migrateSnapshotsAndValidators() public {
        uint256[] memory tokenIDs = new uint256[](4);
        uint256[4] memory tokenIDs_ = _externalStore.getTokenIds();
        for (uint256 i = 0; i < tokenIDs.length; i++) {
            tokenIDs[i] = tokenIDs_[i];
        }
        ////////////// Registering validators /////////////////////////
        address[] memory validatorsAccounts_ = new address[](4);
        validatorsAccounts_[0] = address(
            uint160(0x000000000000000000000000b80d6653f7e5b80dbbe8d0aa9f61b5d72e8028ad)
        );
        validatorsAccounts_[1] = address(
            uint160(0x00000000000000000000000025489d6a720663f7e5253df68948edb302dfdcb6)
        );
        validatorsAccounts_[2] = address(
            uint160(0x000000000000000000000000322e8f463b925da54a778ed597aef41bc4fe4743)
        );
        validatorsAccounts_[3] = address(
            uint160(0x000000000000000000000000adf2a338e19c12298a3007cbea1c5276d1f746e0)
        );

        IValidatorPool(_validatorPoolAddress()).registerValidators(validatorsAccounts_, tokenIDs);
        ///////////////

        // ETHDKG migration
        uint256[] memory validatorIndexes_ = new uint256[](4);
        uint256[4][] memory validatorShares_ = new uint256[4][](4);

        validatorIndexes_[0] = 0x0000000000000000000000000000000000000000000000000000000000000001;
        validatorShares_[0] = [
            0x109f68dde37442959baa4b16498a6fd19c285f9355c23d8eef900876e8536a12,
            0x2c11cec2ce4e17afffcc105f9bd0e646f6274f562f1c93f93545fc22c74a2cdc,
            0x0430024aa1619a117e74425481c44f4628f45af7b389e4d5f84fc41227e1829e,
            0x163cb9abb41800ba5cc1955fd72c0983edc9869a21925006e691d4947451c9fd
        ];

        validatorIndexes_[1] = 0x0000000000000000000000000000000000000000000000000000000000000002;
        validatorShares_[1] = [
            0x13f8a33ff7ef3cb5536b2223195b7b652a3533d309ad3887bcf570c9b1dbe142,
            0x170fe500681ff96e84a6dd7d1e4698f0ad6dd3ef17520a7cac29b29b84f86aa7,
            0x19d29ec38a1d7d8d7284a76b214bf5b818eaccf47cd37ab8bc08c20833e586e9,
            0x27c0f981d1bbc1667ea520341b7fa65fc79815ab8122a814e3714bfdbacd84db
        ];

        validatorIndexes_[2] = 0x0000000000000000000000000000000000000000000000000000000000000003;
        validatorShares_[2] = [
            0x1064dd800716a7e80ed5d40d5563940cd25be4451976a28c237e1426d40eae5a,
            0x0e1c4d27ca7672e0662aaecbc5f2d62ec23e58cf63ffa9b6fabd41d3cff7c927,
            0x106d1f91c4b77d5c9bb485aeea784e9acf0c91702eefb766e94aefc92043a004,
            0x281971fd391a560142b8d796018afc31131a668b2ca6f62b304564d6422bb03f
        ];

        validatorIndexes_[3] = 0x0000000000000000000000000000000000000000000000000000000000000004;
        validatorShares_[3] = [
            0x2228b7dd85ddae13994fa85f42df1833da3b9468a1e65b987142d62f125a9754,
            0x0c5682ae7cd22a3c3daff06ce469f318025845e90254d9d05cecaeba45f445a5,
            0x2cdac99ed82ffc83fc17213e96d56400db23f08d05418936cb352f0e179cf971,
            0x06371376125bb2b96a5e427fac829f5c3919296aac4c42ddc44eb7c773369c2b
        ];

        uint8 validatorCount_ = 4;
        uint256 epoch_ = 1;
        uint256 ethHeight_ = 0x236;
        uint256 sideChainHeight_ = 0;
        uint256[4] memory masterPublicKey_ = [
            0x2c28ce7f0c752e035b68687a8210cceb6068b5034bba9a4a8f2d43e3bbaa8877,
            0x081b33b885370e04cd712601eb860bf821396bdbcd4b089aba0bfe7b1e649dd3,
            0x253adba688741303e0b046632b35289a0d5c7648b414375e4d61a855abc5f0c3,
            0x095ed894617e232df1779101e1d98e177340cb0fc6283cbc437d79a12290c2f1
        ];
        IETHDKG(_ethdkgAddress()).migrateValidators(
            validatorsAccounts_,
            validatorIndexes_,
            validatorShares_,
            validatorCount_,
            epoch_,
            sideChainHeight_,
            ethHeight_,
            masterPublicKey_
        );

        // deposit
        IUtilityToken(_bTokenAddress()).virtualMintDeposit(
            1,
            0xba7809A4114eEF598132461f3202b5013e834CD5,
            500000000000
        );

        // Snapshot migration
        bytes[] memory bClaims_ = new bytes[](5);
        bytes[] memory groupSignatures_ = new bytes[](5);

        bClaims_[0] = abi.encodePacked(
            hex"000000000100040015000000002c01000d00000002010000190000000201000025000000020100003100000002010000031dfcf2fef268ff9956ee399230e9bf1da9dd510d18552b736b3269f4544c01c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470d058c043d927976ecf061b3cdb3a4a0d2de3284fcd69e23733650a1b3ef367533807ec1e085227e7bb99f47db1b118cefdae66f2fbfc66449a4500e9a6a2dab2"
        );
        bClaims_[1] = abi.encodePacked(
            hex"000000000100040015000000003001000d000000020100001900000002010000250000000201000031000000020100009a7a9e6d46b1640392f4444a9cf56d1190fe77fd4a740ee76b0e5f261341d195c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470d29f86626d42e94c88da05e5cec3c29f0fd037f8a9e1fcb6b49a4dd322da317ce4c870a97b5731173a6d17b71740c498ed409e25e28e9077c7f9119af3c28692"
        );
        bClaims_[2] = abi.encodePacked(
            hex"000000000100040015000000003401000d0000000201000019000000020100002500000002010000310000000201000000f396eeda71abea614606937f7fcbd4d704af9ac0556a66687d689497c8da09c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a47033839738f138dbcbb362c3b351c7b7f16041304c75354fb11ae01d3623cc4e146a5a9af572eacd9e40d9f508d077419cc191f542c213d2c204d3251ce88c476b"
        );
        bClaims_[3] = abi.encodePacked(
            hex"000000000100040015000000003801000d0000000201000019000000020100002500000002010000310000000201000000af33d9a061b001d8c1c912b2cf58f5f5bccd81e9c0fac7ac4f256134677a27c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a47099726b1e813baf97a0f88c89c5257358c4ef40c38b515184ea95bb9113587c85a06879b5886d1af4f04773c418b9517db8b410de7fdff0fd9ed47316e4c23e9f"
        );
        bClaims_[4] = abi.encodePacked(
            hex"000000000100040015000000003c01000d000000020100001900000002010000250000000201000031000000020100001923548c43980ec331fa993cb8b90b157f4251fc8c37ba3506d205611af468e8c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a4702df6fa1cfdeabd709149817a42eb2c1e2c18cc06c6b2bbf4a51d825aaa442f3516f8b5f4a60397c0efdd38750282135beff68f4cdff36497574894658e2807ce"
        );

        groupSignatures_[0] = abi.encodePacked(
            hex"2c28ce7f0c752e035b68687a8210cceb6068b5034bba9a4a8f2d43e3bbaa8877081b33b885370e04cd712601eb860bf821396bdbcd4b089aba0bfe7b1e649dd3253adba688741303e0b046632b35289a0d5c7648b414375e4d61a855abc5f0c3095ed894617e232df1779101e1d98e177340cb0fc6283cbc437d79a12290c2f114551b8239e68c2fc16a68bbfbfe2140b718ca279d784074743ce1dcdb134ed10d0a4d630460957d1c50c0e3a8238cafc3985651674ce03e4b91837da6080de6"
        );
        groupSignatures_[1] = abi.encodePacked(
            hex"2c28ce7f0c752e035b68687a8210cceb6068b5034bba9a4a8f2d43e3bbaa8877081b33b885370e04cd712601eb860bf821396bdbcd4b089aba0bfe7b1e649dd3253adba688741303e0b046632b35289a0d5c7648b414375e4d61a855abc5f0c3095ed894617e232df1779101e1d98e177340cb0fc6283cbc437d79a12290c2f11a4d9a0e85b1e265f221c163546d61fcf76b301944368abbfbba42dc56a083ba2ac800dc9a20a25ca95146c65d6c6cddbb299625907c1a057754f70073ec8675"
        );
        groupSignatures_[2] = abi.encodePacked(
            hex"2c28ce7f0c752e035b68687a8210cceb6068b5034bba9a4a8f2d43e3bbaa8877081b33b885370e04cd712601eb860bf821396bdbcd4b089aba0bfe7b1e649dd3253adba688741303e0b046632b35289a0d5c7648b414375e4d61a855abc5f0c3095ed894617e232df1779101e1d98e177340cb0fc6283cbc437d79a12290c2f106e2ca23e60db68ff939899c926fd9d76e40d15b17720bd5d60df4fd9725cd07288ca12870d4b48f441e6a5b1943c8b9c91f0bd28256ab352e77a61d23124dbb"
        );
        groupSignatures_[3] = abi.encodePacked(
            hex"2c28ce7f0c752e035b68687a8210cceb6068b5034bba9a4a8f2d43e3bbaa8877081b33b885370e04cd712601eb860bf821396bdbcd4b089aba0bfe7b1e649dd3253adba688741303e0b046632b35289a0d5c7648b414375e4d61a855abc5f0c3095ed894617e232df1779101e1d98e177340cb0fc6283cbc437d79a12290c2f11bbb68f54eb8ab7b8276432c152909f11ba49cf685c07fadc1e1ba96c1b579ee27002d8fe6bf013b640e1904525645c5f481cc47358330a8b6eb29d019828e33"
        );
        groupSignatures_[4] = abi.encodePacked(
            hex"2c28ce7f0c752e035b68687a8210cceb6068b5034bba9a4a8f2d43e3bbaa8877081b33b885370e04cd712601eb860bf821396bdbcd4b089aba0bfe7b1e649dd3253adba688741303e0b046632b35289a0d5c7648b414375e4d61a855abc5f0c3095ed894617e232df1779101e1d98e177340cb0fc6283cbc437d79a12290c2f1046a12b7354767f6ec2e660540eee970333bfa01e458ee4cd066588d3c4632972e3c60a8f58a5f89b0926ae265b921bed31fc830056980d70e58db642357af02"
        );

        ISnapshots(_snapshotsAddress()).migrateSnapshots(groupSignatures_, bClaims_);
    }
}
