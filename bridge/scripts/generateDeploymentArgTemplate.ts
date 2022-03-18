import { artifacts } from "hardhat";
import {
  getConstructorArgsABI,
  getInitializerArgsABI,
  writeDeploymentArgs,
} from "./lib/deployment/deployArgUtil";
import { getDeploymentList } from "./lib/deployment/deploymentListUtil";
import { ArgData, DeploymentArgs } from "./lib/deployment/deploymentUtil";

async function main() {
  // get an array of all contracts in the artifacts
  const contracts = await getDeploymentList();
  const deploymentArgs: DeploymentArgs = {
    constructor: {},
    initializer: {},
  };
  for (const contract of contracts) {
    // check each contract for a constructor and
    const cArgs: Array<ArgData> = await getConstructorArgsABI(
      contract,
      artifacts
    );
    const iArgs: Array<ArgData> = await getInitializerArgsABI(
      contract,
      artifacts
    );
    if (cArgs.length !== 0) {
      deploymentArgs.constructor[contract] = cArgs;
    }
    if (iArgs.length !== 0) {
      deploymentArgs.initializer[contract] = iArgs;
    }
  }
  await writeDeploymentArgs(deploymentArgs);
}

main()
  .then(() => {
    return 0;
  })
  .catch((error) => {
    console.error(error);
    throw new Error("unexpected error");
  });
