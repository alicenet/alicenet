import toml from "@iarna/toml";
import fs from "fs";
import { BASE_CONFIG_PATH, DEPLOYMENT_CONFIG_PATH, FACTORY_STATE_CONFIG_PATH } from "./constants";
import { FactoryConfig } from "./factoryStateUtil";

export async function readBaseConfig() {
    if (fs.existsSync(BASE_CONFIG_PATH)) {
      let output = fs.readFileSync(BASE_CONFIG_PATH);
      let config = toml.parse(output.toString());
      return config;
    } else {
      throw new Error(`failed to get baseConfig`);
    }
}

export function readDeploymentConfig() {
  if (fs.existsSync(DEPLOYMENT_CONFIG_PATH)) {
    let output = fs.readFileSync(DEPLOYMENT_CONFIG_PATH);
    let config = toml.parse(output.toString());
    return config;
  } else {
    throw new Error(`failed to get baseConfig`);
  }
}

export async function readFactoryStateData() {
  //this output object allows dynamic addition of fields
  let outputObj: any = {};
  //if there is a file or directory at that location
  if (fs.existsSync(FACTORY_STATE_CONFIG_PATH)) {
    let rawData = fs.readFileSync(FACTORY_STATE_CONFIG_PATH);
    const output = toml.parse(rawData.toString("utf8"));
    outputObj = output;
  }
  return outputObj;
}