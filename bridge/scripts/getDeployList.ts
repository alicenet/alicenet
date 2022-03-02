import { DeployList, writeDeploymentList } from "./lib/deploymentListUtil";
import {
  getAllContracts,
  getDeployType,
  InitData,
} from "./lib/deploymentUtils";

async function main() {
  let outputData = <InitData>{
    constructorArgs: {},
    initializerArgs: {},
  };
  //get an array of all contracts in the artifacts
  let contracts = await getAllContracts();
  let output: DeployList = { deployments: [] };
  for (let contract of contracts) {
    let deployType = await getDeployType(contract);
    if (deployType !== undefined) {
      output.deployments.push(contract);
    }
  }
  await writeDeploymentList(output);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
