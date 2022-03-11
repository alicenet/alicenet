import toml from "@iarna/toml";
import fs from "fs";
import { Artifact, Artifacts } from "hardhat/types";
import { DEPLOYMENT_LIST_PATH } from "../constants";
import { readDeploymentList } from "./deploymentConfigUtil";
import { getDeployType, getDeployGroup, getDeployGroupIndex } from "./deploymentUtil";

export type DeploymentList = {
  [key: string]: Array<ContractDeploymentInfo>;
};

export interface ContractDeploymentInfo {
  contract: string;
  index: number;
}

export interface DeploymentGroupIndexList {
  [key: string]: number[];
}

export async function getDeploymentList(listName?: string, usrPath?: string) {
  let lName = listName === undefined ? "default" : listName;
  let path = usrPath === undefined ? DEPLOYMENT_LIST_PATH : usrPath;
  let config: any = await readDeploymentList(path);
  let deploymentList: Array<string>;
  if (config[lName] === undefined) {
    throw new Error(`failed to locate list with name: ${lName}`);
  } else {
    deploymentList = config[lName];
  }
  return deploymentList;
}

export async function transformDeploymentList(deploymentlist: DeploymentList) {
  let list: Array<string> = [];
  for (let group in deploymentlist) {
    for (let item of deploymentlist[group]) {
      list.push(item.contract);
    }
  }
  return list;
}

export async function getSortedDeployList(contracts: Array<string>, artifacts: Artifacts){
  let deploymentList: DeploymentList = {};
  for (let contract of contracts) {
    let deployType: string | undefined = await getDeployType(contract, artifacts);
    let group: string | undefined = await getDeployGroup(contract, artifacts);
    let index = -1
    if (group !== undefined) {
      let indexString: string | undefined = await getDeployGroupIndex(contract, artifacts);
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
  return deploymentList
}

export async function writeDeploymentList(
  list: Array<string>,
  listName?: string,
  usrPath?: string
) {
  let lName = listName === undefined ? "default" : listName;
  let path = usrPath === undefined ? DEPLOYMENT_LIST_PATH : usrPath;
  let isNew = fs.existsSync(path) === false ? true : false;
  let config: any = await readDeploymentList();
  config[lName] = config[lName] === undefined ? {} : config[lName];
  config[lName] = list;
  let data = toml.stringify(config);
  fs.writeFileSync(path, data);
  let output = fs.readFileSync(path).toString().split("\n");
  output.unshift(
    "# WARNING: DO NOT CHANGE THE GENERATED DEFAULT LIST \n# TO ADD A CUSTOM LIST COPY THE FORMAT OF THE DEFAULT LIST WITH DIFFERENT FIELD NAME"
  );
  fs.writeFileSync(path, output.join("\n"));
}
