// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/IValidatorPool.sol";
import "contracts/libraries/ethdkg/ETHDKGStorage.sol";
import "contracts/interfaces/IETHDKGEvents.sol";
import "contracts/utils/ETHDKGUtils.sol";
import "contracts/libraries/math/CryptoLibrary.sol";

/// @custom:salt ETHDKGAccusations
/// @custom:deploy-type deployUpgradeable
contract ETHDKGAccusations is ETHDKGStorage, IETHDKGEvents, ETHDKGUtils {

    constructor() ETHDKGStorage(){
    }

    function accuseParticipantNotRegistered(address[] memory dishonestAddresses) external {
        require(
            _ethdkgPhase == Phase.RegistrationOpen &&
                ((block.number >= _phaseStartBlock + _phaseLength) &&
                    (block.number < _phaseStartBlock + 2 * _phaseLength)),
            "ETHDKG: should be in post-registration accusation phase!"
        );

        uint16 badParticipants = _badParticipants;
        for (uint256 i = 0; i < dishonestAddresses.length; i++) {
            require(
                IValidatorPool(_ValidatorPoolAddress()).isValidator(dishonestAddresses[i]),
                "ETHDKG: Dispute Failed! Dishonest Address is not a validator at the moment!"
            );

            // check if the dishonestParticipant didn't participate in the registration phase,
            // so it doesn't have a Participant object with the latest nonce
            Participant memory dishonestParticipant = _participants[dishonestAddresses[i]];
            require(
                dishonestParticipant.nonce != _nonce,
                "ETHDKG: Dispute failed! dishonestParticipant is participating in this ETHDKG round!"
            );

            // this makes sure we cannot accuse someone twice because a minor fine will be enough to
            // evict the validator from the pool
            IValidatorPool(_ValidatorPoolAddress()).minorSlash(dishonestAddresses[i], msg.sender);
            badParticipants++;
        }
        _badParticipants = badParticipants;
    }

    function accuseParticipantDidNotDistributeShares(address[] memory dishonestAddresses) external {
        require(
            _ethdkgPhase == Phase.ShareDistribution &&
                ((block.number >= _phaseStartBlock + _phaseLength) &&
                    (block.number < _phaseStartBlock + 2 * _phaseLength)),
            "ETHDKG: should be in post-ShareDistribution accusation phase!"
        );

        uint16 badParticipants = _badParticipants;

        for (uint256 i = 0; i < dishonestAddresses.length; i++) {
            require(
                IValidatorPool(_ValidatorPoolAddress()).isValidator(dishonestAddresses[i]),
                "ETHDKG: Dispute Failed! Dishonest Address is not a validator at the moment!"
            );
            Participant memory dishonestParticipant = _participants[dishonestAddresses[i]];
            require(
                dishonestParticipant.nonce == _nonce,
                "ETHDKG: Dispute failed! Dishonest Participant is not participating in this ETHDKG round!"
            );

            require(
                dishonestParticipant.phase != Phase.ShareDistribution,
                "ETHDKG: Dispute failed! Supposed dishonest participant distributed its share in this ETHDKG round!"
            );

            require(
                dishonestParticipant.distributedSharesHash == 0x0,
                "ETHDKG: Dispute failed! Supposed dishonest participant distributed its share in this ETHDKG round!"
            );
            require(
                dishonestParticipant.commitmentsFirstCoefficient[0] == 0 &&
                    dishonestParticipant.commitmentsFirstCoefficient[1] == 0,
                "ETHDKG: Dispute failed! It looks like the supposed dishonest participant had commitments! "
            );

            IValidatorPool(_ValidatorPoolAddress()).minorSlash(dishonestAddresses[i], msg.sender);
            badParticipants++;
        }

        _badParticipants = badParticipants;
    }


    function accuseParticipantDistributedBadShares(
        address dishonestAddress,
        uint256[] memory encryptedShares,
        uint256[2][] memory commitments,
        uint256[2] memory sharedKey,
        uint256[2] memory sharedKeyCorrectnessProof
    ) external {
        // We should allow accusation, even if some of the participants didn't participate
        require(
            (_ethdkgPhase == Phase.DisputeShareDistribution &&
                block.number >= _phaseStartBlock &&
                block.number < _phaseStartBlock + _phaseLength) ||
                (_ethdkgPhase == Phase.ShareDistribution &&
                    (block.number >= _phaseStartBlock + _phaseLength) &&
                    (block.number < _phaseStartBlock + 2 * _phaseLength)),
            "ETHDKG: Dispute failed! Contract is not in dispute phase!"
        );
        require(
            IValidatorPool(_ValidatorPoolAddress()).isValidator(dishonestAddress),
            "ETHDKG: Dispute Failed! Dishonest Address is not a validator at the moment!"
        );

        Participant memory dishonestParticipant = _participants[dishonestAddress];
        Participant memory disputer = _participants[msg.sender];

        require(
            disputer.nonce == _nonce,
            "ETHDKG: Dispute failed! Disputer is not participating in this ETHDKG round!"
        );
        require(
            dishonestParticipant.nonce == _nonce,
            "ETHDKG: Dispute failed! Dishonest Participant is not participating in this ETHDKG round!"
        );

        require(
            dishonestParticipant.phase == Phase.ShareDistribution,
            "ETHDKG: Dispute failed! dishonestParticipant did not distribute shares!"
        );

        require(
            disputer.phase == Phase.ShareDistribution,
            "ETHDKG: Dispute failed! Disputer did not distribute shares!"
        );

        require(
            dishonestParticipant.distributedSharesHash ==
                keccak256(
                    abi.encodePacked(
                        keccak256(abi.encodePacked(encryptedShares)),
                        keccak256(abi.encodePacked(commitments))
                    )
                ),
            "ETHDKG: Dispute failed! Submitted commitments and encrypted shares don't match!"
        );

        require(
            CryptoLibrary.dleq_verify(
                [CryptoLibrary.G1x, CryptoLibrary.G1y],
                disputer.publicKey,
                dishonestParticipant.publicKey,
                sharedKey,
                sharedKeyCorrectnessProof
            ),
            "ETHDKG: Dispute failed! Invalid shared key or proof!"
        );

        // Since all provided data is valid so far, we load the share and use the verified shared
        // key to decrypt the share for the disputer.
        uint256 share;
        if (disputer.index < dishonestParticipant.index) {
            share = encryptedShares[disputer.index - 1];
        } else {
            share = encryptedShares[disputer.index - 2];
        }
        share ^= uint256(keccak256(abi.encodePacked(sharedKey[0], disputer.index)));

        // Verify the share for it's correctness using the polynomial defined by the commitments.
        // First, the polynomial (in group G1) is evaluated at the disputer's idx.
        uint256 x = disputer.index;
        uint256[2] memory result = commitments[0];
        uint256[2] memory tmp = CryptoLibrary.bn128_multiply(
            [commitments[1][0], commitments[1][1], x]
        );
        result = CryptoLibrary.bn128_add([result[0], result[1], tmp[0], tmp[1]]);
        for (uint256 j = 2; j < commitments.length; j++) {
            x = mulmod(x, disputer.index, CryptoLibrary.GROUP_ORDER);
            tmp = CryptoLibrary.bn128_multiply([commitments[j][0], commitments[j][1], x]);
            result = CryptoLibrary.bn128_add([result[0], result[1], tmp[0], tmp[1]]);
        }
        // Then the result is compared to the point in G1 corresponding to the decrypted share.
        // In this case, either the shared value is invalid, so the dishonestAddress
        // should be burned; otherwise, the share is valid, and whoever
        // submitted this accusation should be burned. In any case, someone
        // will have his stake burned.
        tmp = CryptoLibrary.bn128_multiply([CryptoLibrary.G1x, CryptoLibrary.G1y, share]);
        if (result[0] != tmp[0] || result[1] != tmp[1]) {
            IValidatorPool(_ValidatorPoolAddress()).majorSlash(dishonestAddress, msg.sender);
        } else {
            IValidatorPool(_ValidatorPoolAddress()).majorSlash(msg.sender, dishonestAddress);
        }
        _badParticipants++;
    }

    function accuseParticipantDidNotSubmitKeyShares(address[] memory dishonestAddresses) external {
        require(
            _ethdkgPhase == Phase.KeyShareSubmission &&
                (block.number >= _phaseStartBlock + _phaseLength &&
                    block.number < _phaseStartBlock + 2 * _phaseLength),
            "ETHDKG: Dispute failed! Should be in post-KeyShareSubmission phase!"
        );

        uint16 badParticipants = _badParticipants;

        for (uint256 i = 0; i < dishonestAddresses.length; i++) {
            require(
                IValidatorPool(_ValidatorPoolAddress()).isValidator(dishonestAddresses[i]),
                "ETHDKG: Dispute Failed! Dishonest Address is not a validator at the moment!"
            );

            Participant memory dishonestParticipant = _participants[dishonestAddresses[i]];
            require(
                dishonestParticipant.nonce == _nonce,
                "ETHDKG: Dispute failed! Dishonest Participant is not participating in this ETHDKG round!"
            );

            require(
                dishonestParticipant.phase != Phase.KeyShareSubmission,
                "ETHDKG: Dispute failed! Dishonest participant submitted its key shares in this ETHDKG round!"
            );

            require(
                dishonestParticipant.keyShares[0] == 0 && dishonestParticipant.keyShares[1] == 0,
                "ETHDKG: Dispute failed! Dishonest participant submitted its key shares in this ETHDKG round!"
            );

            // evict the validator that didn't submit his shares
            IValidatorPool(_ValidatorPoolAddress()).minorSlash(dishonestAddresses[i], msg.sender);
            badParticipants++;
        }
        _badParticipants = badParticipants;
    }

    function accuseParticipantDidNotSubmitGPKJ(address[] memory dishonestAddresses) external {
        require(
            _ethdkgPhase == Phase.GPKJSubmission &&
                (block.number >= _phaseStartBlock + _phaseLength &&
                    block.number < _phaseStartBlock + 2 * _phaseLength),
            "ETHDKG: Dispute Failed! Should be in post-GPKJSubmission phase!"
        );

        uint16 badParticipants = _badParticipants;

        for (uint256 i = 0; i < dishonestAddresses.length; i++) {
            require(
                IValidatorPool(_ValidatorPoolAddress()).isValidator(dishonestAddresses[i]),
                "ETHDKG: Dispute Failed! Dishonest Address is not a validator at the moment!"
            );
            Participant memory dishonestParticipant = _participants[dishonestAddresses[i]];
            require(
                dishonestParticipant.nonce == _nonce,
                "ETHDKG: Dispute failed! Dishonest Participant is not participating in this ETHDKG round!"
            );

            require(
                dishonestParticipant.phase != Phase.GPKJSubmission,
                "ETHDKG: Dispute failed! Dishonest participant did participate in this GPKj submission!"
            );

            // todo: being paranoic, check if we need this or if it's expensive
            require(
                dishonestParticipant.gpkj[0] == 0 &&
                    dishonestParticipant.gpkj[1] == 0 &&
                    dishonestParticipant.gpkj[2] == 0 &&
                    dishonestParticipant.gpkj[3] == 0,
                "ETHDKG: Dispute failed! It looks like the dishonestParticipant distributed its GPKJ!"
            );

            IValidatorPool(_ValidatorPoolAddress()).minorSlash(dishonestAddresses[i], msg.sender);
            badParticipants++;
        }

        _badParticipants = badParticipants;
    }

    function accuseParticipantSubmittedBadGPKJ(
        address[] memory validators,
        bytes32[] memory encryptedSharesHash,
        uint256[2][][] memory commitments,
        address dishonestAddress
    ) external {
        // We should allow accusation, even if some of the participants didn't participate
        require(
            (_ethdkgPhase == Phase.DisputeGPKJSubmission &&
                block.number >= _phaseStartBlock &&
                block.number < _phaseStartBlock + _phaseLength) ||
                (_ethdkgPhase == Phase.GPKJSubmission &&
                    (block.number >= _phaseStartBlock + _phaseLength) &&
                    (block.number < _phaseStartBlock + 2 * _phaseLength)),
            "ETHDKG: Dispute Failed! Should be in post-GPKJSubmission phase!"
        );

        require(
            IValidatorPool(_ValidatorPoolAddress()).isValidator(dishonestAddress),
            "ETHDKG: Dispute Failed! Dishonest Address is not a validator at the moment!"
        );

        Participant memory dishonestParticipant = _participants[dishonestAddress];
        Participant memory disputer = _participants[msg.sender];

        require(
            dishonestParticipant.nonce == _nonce &&
                dishonestParticipant.phase == Phase.GPKJSubmission,
            "ETHDKG: Dispute Failed! Dishonest participant didn't submit his GPKJ for this round!"
        );

        require(
            disputer.nonce == _nonce && disputer.phase == Phase.GPKJSubmission,
            "ETHDKG: Dispute Failed! Disputer didn't submit his GPKJ for this round!"
        );

        uint16 badParticipants = _badParticipants;
        // n is total _participants;
        // t is threshold, so that t+1 is BFT majority.
        uint256 numParticipants = IValidatorPool(_ValidatorPoolAddress()).getValidatorsCount() + badParticipants;
        uint256 threshold = _getThreshold(numParticipants);

        // Begin initial check
        ////////////////////////////////////////////////////////////////////////
        // First, check length of things
        require(
            (validators.length == numParticipants) &&
                (encryptedSharesHash.length == numParticipants) &&
                (commitments.length == numParticipants),
            "ETHDKG: Dispute Failed! Invalid submission of arguments!"
        );
        {
            uint256 bitMap = 0;
            uint256 nonce = _nonce;
            // Now, ensure sub-arrays are the correct length as well
            for (uint256 k = 0; k < numParticipants; k++) {
                require(
                    commitments[k].length == threshold + 1,
                    "ETHDKG: Dispute Failed! Invalid number of commitments provided!"
                );

                bytes32 commitmentsHash = keccak256(abi.encodePacked(commitments[k]));
                Participant memory participant = _participants[validators[k]];
                require(
                    participant.nonce == nonce &&
                        participant.index <= type(uint8).max &&
                        !_isBitSet(bitMap, uint8(participant.index)),
                    "ETHDKG: Dispute Failed! Invalid or duplicated participant address!"
                );

                require(
                    participant.distributedSharesHash ==
                        keccak256(abi.encodePacked(encryptedSharesHash[k], commitmentsHash)),
                    "ETHDKG: Dispute Failed! Invalid shares or commitments!"
                );
                bitMap = _setBit(bitMap, uint8(participant.index));
            }
        }

        ////////////////////////////////////////////////////////////////////////
        // End initial check

        // Info for looping computation
        uint256 pow;
        uint256[2] memory gpkjStar;
        uint256[2] memory tmp;
        uint256 idx;

        // Begin computation loop
        //
        // We remember
        //
        //      F_i(x) = C_i0 * C_i1^x * C_i2^(x^2) * ... * C_it^(x^t)
        //             = Prod(C_ik^(x^k), k = 0, 1, ..., t)
        //
        // We now compute gpkj*. We have
        //
        //      gpkj* = Prod(F_i(j), i)
        //            = Prod( Prod(C_ik^(j^k), k = 0, 1, ..., t), i)
        //            = Prod( Prod(C_ik^(j^k), i), k = 0, 1, ..., t)    // Switch order
        //            = Prod( [Prod(C_ik, i)]^(j^k), k = 0, 1, ..., t)  // Move exponentiation outside
        //
        // More explicitly, we have
        //
        //      gpkj* =  Prod(C_i0, i)        *
        //              [Prod(C_i1, i)]^j     *
        //              [Prod(C_i2, i)]^(j^2) *
        //                  ...
        //              [Prod(C_it, i)]^(j^t) *
        //
        ////////////////////////////////////////////////////////////////////////
        // Add constant terms
        gpkjStar = commitments[0][0]; // Store initial constant term
        for (idx = 1; idx < numParticipants; idx++) {
            gpkjStar = CryptoLibrary.bn128_add(
                [gpkjStar[0], gpkjStar[1], commitments[idx][0][0], commitments[idx][0][1]]
            );
        }

        // Add linear term
        tmp = commitments[0][1]; // Store initial linear term
        pow = dishonestParticipant.index;
        for (idx = 1; idx < numParticipants; idx++) {
            tmp = CryptoLibrary.bn128_add(
                [tmp[0], tmp[1], commitments[idx][1][0], commitments[idx][1][1]]
            );
        }
        tmp = CryptoLibrary.bn128_multiply([tmp[0], tmp[1], pow]);
        gpkjStar = CryptoLibrary.bn128_add([gpkjStar[0], gpkjStar[1], tmp[0], tmp[1]]);

        // Loop through higher order terms
        for (uint256 k = 2; k <= threshold; k++) {
            tmp = commitments[0][k]; // Store initial degree k term
            // Increase pow by factor
            pow = mulmod(pow, dishonestParticipant.index, CryptoLibrary.GROUP_ORDER);
            for (idx = 1; idx < numParticipants; idx++) {
                tmp = CryptoLibrary.bn128_add(
                    [tmp[0], tmp[1], commitments[idx][k][0], commitments[idx][k][1]]
                );
            }
            tmp = CryptoLibrary.bn128_multiply([tmp[0], tmp[1], pow]);
            gpkjStar = CryptoLibrary.bn128_add([gpkjStar[0], gpkjStar[1], tmp[0], tmp[1]]);
        }
        ////////////////////////////////////////////////////////////////////////
        // End computation loop

        // We now have gpkj*; we now verify.
        uint256[4] memory gpkj = dishonestParticipant.gpkj;
        bool isValid = CryptoLibrary.bn128_check_pairing(
            [
                gpkjStar[0],
                gpkjStar[1],
                CryptoLibrary.H2xi,
                CryptoLibrary.H2x,
                CryptoLibrary.H2yi,
                CryptoLibrary.H2y,
                CryptoLibrary.G1x,
                CryptoLibrary.G1y,
                gpkj[0],
                gpkj[1],
                gpkj[2],
                gpkj[3]
            ]
        );
        if (!isValid) {
            IValidatorPool(_ValidatorPoolAddress()).majorSlash(dishonestAddress, msg.sender);
        } else {
           IValidatorPool(_ValidatorPoolAddress()).majorSlash(msg.sender, dishonestAddress);
        }
        badParticipants++;
        _badParticipants = badParticipants;
    }
}
