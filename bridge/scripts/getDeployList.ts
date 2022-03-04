import { DeploymentList, writeDeploymentList, DeploymentGroupIndexList, ContractDeploymentInfo, transformDeploymentList } from "./lib/deploymentListUtil";
import {
  getAllContracts,
  getDeployGroup,
  getDeployGroupIndex,
  getDeployType,
  InitData,
} from "./lib/deploymentUtil";

async function main() {
  let outputData = <InitData>{
    constructorArgs: {},
    initializerArgs: {},
  };
  //get an array of all contracts in the artifacts
  let contracts = await getAllContracts();
  let deploymentList: DeploymentList = {};
  for (let contract of contracts) {
    let deployType: string | undefined = await getDeployType(contract);
    let group: string | undefined = await getDeployGroup(contract);
    let index = -1
    if (group !== undefined) {
      let indexString: string | undefined = await getDeployGroupIndex(contract);
      if (indexString === undefined) {
        throw "If deploy-group-index is specified a deploy-group-index also should be!"
      }
      try {
        index = parseInt(indexString)
      } catch(error) {
        throw `Failed to convert deploy-group-index for contract ${contract}! deploy-group-index should be an integer!`
      }
    } else {
      group = "general"
    }
    if(deployType !== undefined) {
      if (deploymentList[group] === undefined) {
        deploymentList[group] = []
      }
      deploymentList[group].push({contract, index})
    }    
  }
  for (let key in deploymentList) {
    if (key !== "general") {
      deploymentList[key].sort((contractA, contractB) => {
        return contractA.index - contractB.index
      })    
    }
  }
  let list = await transformDeploymentList(deploymentList);
  await writeDeploymentList(list);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
