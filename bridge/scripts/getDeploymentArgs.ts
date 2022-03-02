import fs from "fs";
import { env } from "./lib/constants";
import {
  getAllContracts,
  getDeployType,
  getConstructorArgsABI,
  ArgData,
  getInitializerArgsABI,
  parseArgsArray,
  InitData,
} from "./lib/deploymentUtils";

async function main() {
  let outputData = <InitData>{
    constructorArgs: {},
    initializerArgs: {},
  };
  //get an array of all contracts in the artifacts
  let contracts = await getAllContracts();

  for (let contract of contracts) {
    let deployType = await getDeployType(contract);
    if (deployType !== undefined) {
      //check each contract for a constructor and
      let cArgs: Array<ArgData> = await getConstructorArgsABI(contract);
      let iArgs: Array<ArgData> = await getInitializerArgsABI(contract);
      let cTemplate = parseArgsArray(cArgs);
      let iTemplate = parseArgsArray(iArgs);
      if (cArgs.length != 0) {
        outputData.constructorArgs[contract] = cTemplate;
      }
      if (iArgs.length != 0) {
        outputData.initializerArgs[contract] = iTemplate;
      }
    }
  }

  fs.writeFileSync(
    `./deployments/${env()}/deploymentArgsTemplate.json`,
    JSON.stringify(outputData, null, 2)
  );
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
