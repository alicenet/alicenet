import { BigNumberish, BytesLike } from "ethers";
import fs from "fs";
import { env } from "./constants";
export type FactoryData = {
  address: string;
  gas?: number;
};

export type DeployCreateData = {
  name: string;
  address: string;
  factoryAddress: string;
  gas: number;
  constructorArgs?: any;
};
export type MetaContractData = {
  metaAddress: string;
  salt: string;
  templateName: string;
  templateAddress: string;
  factoryAddress: string;
  gas: number;
  initCallData: string;
};
export type TemplateData = {
  name: string;
  address: string;
  factoryAddress: string;
  gas: number;
  constructorArgs?: string;
};

export interface FactoryConfig {
  [key: string]: any;
}
export type ProxyData = {
  proxyAddress: BytesLike;
  salt: BytesLike;
  logicName: string;
  logicAddress: BytesLike;
  factoryAddress: BytesLike;
  gas: BigNumberish;
  initCallData?: BytesLike;
};

export async function getDefaultFactoryAddress(): Promise<string> {
  //fetch whats in the factory config file
  let config = await readFactoryStateData();
  return config.defaultFactoryData.address;
}

export async function readFactoryStateData() {
  //this output object allows dynamic addition of fields
  let outputObj: FactoryConfig = {};
  //if there is a file or directory at that location
  if (fs.existsSync(`./deployments/${env()}/factoryState.json`)) {
    let rawData = fs.readFileSync(`./deployments/${env()}/factoryState.json`);
    const output = await JSON.parse(rawData.toString("utf8"));
    outputObj = output;
  }
  return outputObj;
}

async function writeFactoryConfig(
  newFactoryConfig: FactoryConfig,
  lastFactoryConfig?: FactoryConfig
) {
  let jsonString = JSON.stringify(newFactoryConfig, null, 2);
  if (lastFactoryConfig !== undefined) {
    let date = new Date();
    let timestamp = date.toUTCString().replace(" ", "_").replace(",", "");
    if (!fs.existsSync(`./deployments/${env()}/archive`)) {
      fs.mkdirSync(`./deployments/${env()}/archive`);
    }
    fs.writeFileSync(
      `./deployments/${env()}/archive/${timestamp}_factoryState.json`,
      jsonString
    );
  }
  fs.writeFileSync(`./deployments/${env()}/factoryState.json`, jsonString);
}
async function getLastConfig(config: FactoryConfig) {
  if (
    config.defaultFactoryData !== undefined &&
    Object.keys(config.defaultFactoryData).length > 0
  ) {
    return config;
  } else {
    return undefined;
  }
}

export async function updateDefaultFactoryData(input: FactoryData) {
  let state = await readFactoryStateData();
  let lastConfig = await getLastConfig(state);
  state.defaultFactoryData = input;
  await writeFactoryConfig(state, lastConfig);
}

export async function updateDeployCreateList(data: DeployCreateData) {
  //fetch whats in the factory config file
  //It is safe to use as
  let config = await readFactoryStateData();
  config.rawDeployments = config.rawDeployments === undefined ? [] : config.rawDeployments;
  config.rawDeployments.push(data)
  // write new data to config file
  await writeFactoryConfig(config);
}

export async function updateTemplateList(data: TemplateData) {
  //fetch whats in the factory config file
  let config = await readFactoryStateData();
  config.templates = config.templates === undefined ? [] : config.templates;
  config.templates.push(data);
  // write new data to config file
  await writeFactoryConfig(config);
}

/**
 * @description pulls in the factory config data and adds proxy data
 * to the proxy array
 * @param data object that contains the proxies
 * logic contract name, address, and proxy address
 */
export async function updateProxyList(data: ProxyData) {
  //fetch whats in the factory config file
  let config = await readFactoryStateData();
  config.proxies = config.proxies === undefined ? [] : config.proxies;
  config.proxies.push(data);
  // write new data to config file
  await writeFactoryConfig(config);
}

export async function updateMetaList(data: MetaContractData) {
  //fetch whats in the factory config file
  let config = await readFactoryStateData();
  config.staticContracts = config.staticContracts === undefined ? [] : config.staticContracts;
  config.staticContracts.push(data);
  // write new data to config file
  await writeFactoryConfig(config);
}
