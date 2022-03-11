import { artifacts } from "hardhat";
import { DeploymentList, writeDeploymentList, DeploymentGroupIndexList, ContractDeploymentInfo, transformDeploymentList, getSortedDeployList } from "./lib/deployment/deploymentListUtil";
import { InitData, getAllContracts, getDeployType, getDeployGroup, getDeployGroupIndex } from "./lib/deployment/deploymentUtil";


async function main() {
  //get an array of all contracts in the artifacts
  let contracts = await getAllContracts(artifacts);
  let deploymentList: DeploymentList = await getSortedDeployList(contracts, artifacts);
  let list = await transformDeploymentList(deploymentList);
  await writeDeploymentList(list);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
