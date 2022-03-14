import toml from "@iarna/toml";
import fs from "fs";
import { DEPLOYMENT_ARG_PATH, DEPLOYMENT_LIST_PATH, FACTORY_STATE_PATH } from "../constants";
import { FactoryConfig } from "./factoryStateUtil";

export async function readDeploymentList(usrPath?: string) {
  return await readTOML(DEPLOYMENT_LIST_PATH, usrPath)
}

export async function readDeploymentArgs(usrPath?: string) {
  return await readTOML(DEPLOYMENT_ARG_PATH, usrPath)
}

export async function readFactoryState(usrPath?: string) {
  return await readTOML(FACTORY_STATE_PATH, usrPath)
}


export async function readTOML(defaultPath: string, usrPath?: string){
  let path = usrPath === undefined ? defaultPath : usrPath;
  let outputObj: any = {};
  if (fs.existsSync(path)) {
    let rawData = fs.readFileSync(path);
    const output = toml.parse(rawData.toString("utf8"));
    outputObj = output;
  }
  return outputObj;
}