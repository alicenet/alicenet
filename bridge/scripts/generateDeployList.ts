import { artifacts } from "hardhat";
import {
  DeploymentList,
  getSortedDeployList,
  transformDeploymentList,
  writeDeploymentList,
} from "./lib/deployment/deploymentListUtil";
import { getAllContracts } from "./lib/deployment/deploymentUtil";

async function main() {
  // get an array of all contracts in the artifacts
  const contracts = await getAllContracts(artifacts);
  const deploymentList: DeploymentList = await getSortedDeployList(
    contracts,
    artifacts
  );
  const list = await transformDeploymentList(deploymentList);
  await writeDeploymentList(list);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    throw new Error("unexpected error");
  });
