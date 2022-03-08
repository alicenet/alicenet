import toml from "@iarna/toml";
import fs from "fs";
import { ethers } from "hardhat";
import { readBaseConfig } from "./baseConfigUtil";
import { BASE_CONFIG_PATH } from "./constants";

export type DeploymentList = {
  [key: string]: Array<ContractDeploymentInfo>;
}

export interface ContractDeploymentInfo {
  contract : string
  index: number
}

export interface DeploymentGroupIndexList {
  [key: string]: number[];
}

export async function getDeploymentList() {
  let config: any = await readBaseConfig();
  let deploymentList: Array<string> = config.deploymentList; 
  return deploymentList;
}
export async function transformDeploymentList(deploymentlist: DeploymentList) {
  let list: Array<string> = [];
  for( let group in deploymentlist){
    for(let item of deploymentlist[group]){
      list.push(item.contract)
    }
  }
  return list;
}

export async function writeDeploymentList(list: Array<string>) {
  let config:any = await readBaseConfig();
  if (config !== undefined) {
    config.deploymentList = list;
    let data = toml.stringify(config);
    fs.writeFileSync(BASE_CONFIG_PATH, data);
  } else {
    throw new Error(`deployment list not found`);
  }
}
