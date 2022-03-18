import {task} from "hardhat/config";
import fs from "fs";


interface DeployList {
    deployments: string[];
}

task
("generate-immutable-auth-contract", "Generate contracts")
    .addOptionalParam("path", "file that contains the list of contracts")
    .setAction(async ({path}) => {

        let contractNameSaltMap = []
        let deployList : DeployList

        // file not defined,
        if (path == null) {
            console.log("UNDEFINED")
            console.log(process.cwd())
            const defaultPathForContracts = 'deployments/dev/deployList.json'
            // If deployments/dev/deployList.json exists then go on, otherwise return error with suggestion to run npx hardhat run ./scripts/getDeployList.ts
            if (fs.existsSync(defaultPathForContracts)) {
                try {
                    const jsonContent = fs.readFileSync(defaultPathForContracts).toString()
                    deployList = JSON.parse(jsonContent);
                } catch (err) {
                    throw new Error("Error parsing JSON file.");
                }

                for (let contract of deployList.deployments){
                    contractNameSaltMap.push([contract,contract])
                }
                console.log(contractNameSaltMap)

            } else {
                throw new Error(
                    "The default 'deployList.json' file do not exists. " +
                    "Run the following command and then try again \n\n npx hardhat run ./scripts/getDeployList.ts\n\n "
                );
            }
        } else {
            console.log("defined  -  ", path)
        }

        //
        // contractNameSaltMap = [
        //     ["ValidatorNFT", "ValidatorNFT"],
        //     ['MadToken', 'MadToken'],
        //     ['StakeNFT', 'StakeNFT'],
        //     ['MadByte', 'MadByte'],
        //     ['Governance', 'Governance'],
        //     ['ValidatorPool', 'ValidatorPool'],
        //     ['ETHDKG', 'ETHDKG'],
        //     ['ETHDKGAccusations', 'ETHDKGAccusations'],
        //     ['Snapshots', 'Snapshots'],
        //     ['ETHDKGPhases', 'ETHDKGPhases'],
        //     ['StakeNFTLP', 'StakeNFTLP'],
        //     ['Foundation', 'Foundation'],
        // ];
        //
        // console.log(templateSalt)
        // for (const contractNameSalt of contractNameSaltMap) {
        //     let name = contractNameSalt[0];
        //     let salt = contractNameSalt[1];
        //
        //     const saltEncoded = Buffer.from(salt, 'utf8');
        //     const zeroByteBuffer = new Buffer(32)
        //     zeroByteBuffer.fill("00", 0, zeroByteBuffer.length, 'hex')
        //     const bufferArray = [saltEncoded, zeroByteBuffer];
        //     const saltBuffer = Buffer.concat(bufferArray).slice(0, 32);
        //     salt = Buffer.from(saltBuffer).toString('hex')
        //
        //     let c = new Contract(name, salt);
        //     let render = templateContract(c)
        //     console.log(render)
        // }
    });

export class Contract {
    name: string;
    salt: string;

    constructor(name: string, salt: string) {
        this.name = name;
        this.salt = salt;
    }
}

let templateSalt = `
// SPDX-License-Identifier: MIT-open-group
pragma solidity 0.8.11;

import "./DeterministicAddress.sol";

abstract contract ImmutableFactory is DeterministicAddress {

    address private immutable _factory;

    constructor(address factory_) {
        _factory = factory_;
    }

    modifier onlyFactory() {
        require(msg.sender == _factory, "onlyFactory");
        _;
    }

    function _factoryAddress() internal view returns (address) {
        return _factory;
    }

}
`

function templateContract(contract: Contract): string {

    return `
    abstract contract immutable${contract.name} is ImmutableFactory {
    
        address private immutable _${contract.name};
    
        constructor() {
            _${contract.name} = getMetamorphicContractAddress(0x${contract.salt}, _factoryAddress());
        }
    
        modifier only${contract.name}() {
            require(msg.sender == _${contract.name}, "only${contract.name}");
            _;
        }
    
        function _${contract.name}Address() internal view returns(address) {
            return _${contract.name};
        }
    
        function _saltFor${contract.name}() internal pure returns(bytes32) {
            return 0x${contract.salt};
        }
    }
    `;
}

