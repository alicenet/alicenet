import fs from "fs";
import { env } from "./constants";
export type DeployList = {
  deployments: Array<string>;
};

export async function readDeploymentList() {
  //this output object allows dynamic addition of fields
  let outputObj = <DeployList>{};
  //if there is a file or directory at that location
  if (fs.existsSync(`./deployments/${env()}/deployList.json`)) {
    let rawData = fs.readFileSync(`./deployments/${env()}/deployList.json`);
    const output = await JSON.parse(rawData.toString("utf8"));
    outputObj = output;
  }
  return outputObj;
}

export async function getDeploymentList() {
  let deployList = await readDeploymentList();
  return deployList.deployments;
}

export async function writeDeploymentList(newFactoryConfig: DeployList) {
  let jsonString = JSON.stringify(newFactoryConfig, null, 2);
  if (!fs.existsSync(`./deployments/`)) {
    fs.mkdirSync(`./deployments/`);
  }
  if (!fs.existsSync(`./deployments/${env()}/`)) {
    fs.mkdirSync(`./deployments/${env()}/`);
  }
  fs.writeFileSync(`./deployments/${env()}/deployList.json`, jsonString);
}
