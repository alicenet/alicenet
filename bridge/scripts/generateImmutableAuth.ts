import {task} from "hardhat/config";
import {getDeploymentList} from "./lib/deployment/deploymentListUtil";
import {HardhatRuntimeEnvironment} from "hardhat/types";
import fs from "fs";

// Generates ImmutableAuth.sol contract. Possible inputs
//
// $ npx hardhat generate-immutable-auth-contract
// $ npx hardhat generate-immutable-auth-contract --input ./path/to/folder/
// $ npx hardhat generate-immutable-auth-contract --output ./
// $ npx hardhat generate-immutable-auth-contract --input ./path/to/folder/ --output ./
//
//
// If you are linking a custom Deployment list file it should follow the following pattern
//
// deploymentList = [
//     "path/to/contract/ContractName.sol:ContractName",
//      ...
// ]

task
("generate-immutable-auth-contract", "Generate contracts")
    .addOptionalParam("input", "path to the folder containing the deploymentList file", undefined)
    .addOptionalParam("output", "path to save the generated `ImmutableAuth.sol`", undefined)
    .setAction(async ({input, output}, hre) => {

        let contractNameSaltMap = []
        let contract = ""

        const contracts = await getDeploymentList(input);
        for (let i = 0; i < contracts.length; i++) {
            const fullyQualifiedName = contracts[i];
            let contractName = fullyQualifiedName.split(":")[1]
            let salt = await getSalt(contractName, hre)
            contractNameSaltMap.push([contractName, salt])
        }

        contract += templateSalt
        for (const contractNameSalt of contractNameSaltMap) {
            let name = contractNameSalt[0];
            let salt = contractNameSalt[1];

            const saltEncoded = Buffer.from(salt, 'utf8');
            const zeroByteBuffer = new Buffer(32)
            zeroByteBuffer.fill("00", 0, zeroByteBuffer.length, 'hex')
            const bufferArray = [saltEncoded, zeroByteBuffer];
            const saltBuffer = Buffer.concat(bufferArray).slice(0, 32);
            salt = Buffer.from(saltBuffer).toString('hex')

            let c = new Contract(name, salt);
            contract += templateContract(c)
        }

        if (output === undefined) {
            output = "contracts/utils/"
        } else {
            checkUserDirPath(output)
        }
        fs.writeFileSync(output + "ImmutableAuth.sol", contract);

        console.log("\nDone. Don't forget to run the lint command \n\n $ npm run lint-solidity\n\n or \n\n $ npm run linter \n\n")
    });


export class Contract {
    name: string;
    salt: string;

    constructor(name: string, salt: string) {
        this.name = name;
        this.salt = salt;
    }
}

let templateSalt = `// SPDX-License-Identifier: MIT-open-group
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

function extractPath(qualifiedName: string) {
    return qualifiedName.split(":")[0];
}

async function checkUserDirPath(path: string) {
    if (path !== undefined) {
        if (!fs.existsSync(path)) {
            console.log("Creating Folder at" + path + " since it didn't exist before!");
            fs.mkdirSync(path);
        }
        if (fs.statSync(path).isFile()) {
            throw new Error("outputFolder path should be to a directory not a file");
        }
    }
}

async function getFullyQualifiedName(
    contractName: string,
    hre: HardhatRuntimeEnvironment
) {
    const artifactPaths = await hre.artifacts.getAllFullyQualifiedNames();
    for (let i = 0; i < artifactPaths.length; i++) {
        if (artifactPaths[i].split(":").length > 0 && artifactPaths[i].split(":")[1] === contractName) {
            return String(artifactPaths[i]);
        }
    }
}

async function getSalt(
    contractName: string,
    hre: HardhatRuntimeEnvironment
): Promise<string> {
    const qualifiedName: any = await getFullyQualifiedName(contractName, hre);
    const buildInfo = await hre.artifacts.getBuildInfo(qualifiedName);
    let contractOutput: any;
    let devdoc: any;
    let salt: string = "";
    if (buildInfo !== undefined) {
        const path = extractPath(qualifiedName);
        contractOutput = buildInfo.output.contracts[path][contractName];
        devdoc = contractOutput.devdoc;
        salt = devdoc["custom:salt"];
        return salt;
    } else {
        console.error("missing salt");
    }
    return salt;
}
