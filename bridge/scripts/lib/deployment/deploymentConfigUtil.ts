import toml from "@iarna/toml";
import fs from "fs";
import {
  DEPLOYMENT_ARG_PATH,
  DEPLOYMENT_LIST_PATH,
  FACTORY_STATE_PATH,
} from "../constants";

export async function readDeploymentList(usrPath?: string) {
  return await readTOML(DEPLOYMENT_LIST_PATH, usrPath);
}

export async function readDeploymentArgs(usrPath?: string) {
  return await readTOML(DEPLOYMENT_ARG_PATH, usrPath);
}

export async function readFactoryState(usrPath?: string) {
  const path =
    usrPath === undefined
      ? FACTORY_STATE_PATH
      : usrPath.replace(/\/+$/, "") + "/factoryState";
  return await readTOML(FACTORY_STATE_PATH, path);
}

export async function readTOML(defaultPath: string, usrPath?: string) {
  const path = usrPath === undefined ? defaultPath : usrPath;
  let outputObj: any = {};
  if (fs.existsSync(path)) {
    const rawData = fs.readFileSync(path);
    const output = toml.parse(rawData.toString("utf8"));
    outputObj = output;
  }
  return outputObj;
}
