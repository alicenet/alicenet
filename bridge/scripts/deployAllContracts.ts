import { STATIC_DEPLOYMENT, UPGRADEABLE_DEPLOYMENT } from "./lib/constants";
import { getDeploymentList } from "./lib/deployment/deploymentListUtil";
import { getDeployType, deployUpgradeableProxy, deployStatic, deployFactory} from "./lib/deployment/deploymentUtil";
import {artifacts, ethers, run} from "hardhat"

async function main() {
  //deploy the factory first
  await deployFactory(run);
  //get an array of all contracts in the artifacts
  let contracts = await getDeploymentList();
  //let contracts = ["src/tokens/periphery/validatorPool/Snapshots.sol:Snapshots"]
  for (let i = 0; i < contracts.length; i++) {
    let fullyQualifiedName = contracts[i];
    //check the contract for the @custom:deploy-type tag
    let deployType = await getDeployType(fullyQualifiedName, artifacts);
    switch (deployType) {
      case STATIC_DEPLOYMENT:
        await deployStatic(fullyQualifiedName, artifacts, ethers, run);
        break;
      case UPGRADEABLE_DEPLOYMENT:
        await deployUpgradeableProxy(fullyQualifiedName, artifacts, ethers, run);
        break;
      default:
        break;
    }
  }
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
