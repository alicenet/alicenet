import { artifacts, ethers, run } from "hardhat";
import { STATIC_DEPLOYMENT, UPGRADEABLE_DEPLOYMENT } from "./lib/constants";
import { getDeploymentList } from "./lib/deployment/deploymentListUtil";
import {
  deployFactory,
  getDeployMetaArgs,
  getDeployUpgradeableProxyArgs,
  getDeployType,
} from "./lib/deployment/deploymentUtil";

async function main() {
  // deploy the factory first
  await deployFactory(run);
  // get an array of all contracts in the artifacts
  const contracts = await getDeploymentList();
  // let contracts = ["src/tokens/periphery/validatorPool/Snapshots.sol:Snapshots"]
  for (let i = 0; i < contracts.length; i++) {
    const fullyQualifiedName = contracts[i];
    // check the contract for the @custom:deploy-type tag
    const deployType = await getDeployType(fullyQualifiedName, artifacts);
    switch (deployType) {
      case STATIC_DEPLOYMENT:
        await getDeployMetaArgs(fullyQualifiedName, artifacts, ethers, run);
        break;
      case UPGRADEABLE_DEPLOYMENT:
        await getDeployUpgradeableProxyArgs(
          fullyQualifiedName,
          artifacts,
          ethers,
          run
        );
        break;
      default:
        break;
    }
  }
}

main()
  .then(() => {
    return 0;
  })
  .catch((error) => {
    console.error(error);
    throw new Error("unexpected error");
  });
