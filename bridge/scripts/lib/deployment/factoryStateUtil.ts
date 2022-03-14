import { BytesLike, BigNumberish, ContractReceipt } from "ethers";
import fs from "fs";
import { readDeploymentArgs, readFactoryState } from "./deploymentConfigUtil";
import { DEPLOYMENT_ARG_PATH, FACTORY_STATE_PATH } from "../constants";
import toml from "@iarna/toml";

export type FactoryData = {
  address: string;
  owner?: string;
  gas?: number;
};

export type DeployCreateData = {
  name: string;
  address: string;
  factoryAddress: string;
  gas: number;
  constructorArgs?: any;
  receipt: ContractReceipt;
};
export type MetaContractData = {
  metaAddress: string;
  salt: string;
  templateName: string;
  templateAddress: string;
  factoryAddress: string;
  gas: number;
  initCallData: string;
  receipt: ContractReceipt;
};
export type TemplateData = {
  name: string;
  address: string;
  factoryAddress: string;
  gas: number;
  receipt: ContractReceipt;
  constructorArgs?: string;
};

export interface FactoryConfig {
  [key: string]: any;
}
export type ProxyData = {
  proxyAddress: string;
  salt: BytesLike;
  logicName?: string;
  logicAddress?: string;
  factoryAddress: string;
  gas?: BigNumberish;
  receipt?: ContractReceipt;
  initCallData?: BytesLike;
};

export async function getDefaultFactoryAddress(network:string, fileName?:string): Promise<string> {
  let path = fileName === undefined ? FACTORY_STATE_PATH : FACTORY_STATE_PATH.split("factoryState")[0] + fileName;
  //fetch whats in the factory config file
  let config = await readFactoryState(path);
  return config[network].defaultFactoryAddress;
}

async function writeFactoryState(network: string, fieldName: string, fieldData: any, fileName?:string){
  let path = fileName === undefined ? FACTORY_STATE_PATH : FACTORY_STATE_PATH.split("factoryState")[0] + fileName;
  let factoryStateConfig
  if(fs.existsSync(path)){
    factoryStateConfig = await readFactoryState(path);
    factoryStateConfig[network] = factoryStateConfig[network] === undefined ? {} : factoryStateConfig[network];
    factoryStateConfig[network][fieldName] = fieldData;
  }else{
    factoryStateConfig = {
      [network]: {
        [fieldName]: fieldData
      }
    }
  }
  let data = toml.stringify(factoryStateConfig)
  fs.writeFileSync(path, data);
}

export async function updateDefaultFactoryData(network: string, data: FactoryData, fileName?:string) {
  let path = fileName === undefined ? FACTORY_STATE_PATH : FACTORY_STATE_PATH.split("factoryState")[0] + fileName;
  await writeFactoryState(network, "defaultFactoryData", data, fileName)
  await writeFactoryState(network, "defaultFactoryAddress", data.address, fileName)
}

export async function updateDeployCreateList(network: string, data: DeployCreateData, fileName?:string) {
  
  await updateList(network, "rawDeployments", data)
}

export async function updateTemplateList(network: string, data: TemplateData, fileName?:string) {
  
  await updateList(network, "templates", data);
}

/**
 * @description pulls in the factory config data and adds proxy data
 * to the proxy array
 * @param data object that contains the proxies
 * logic contract name, address, and proxy address
 */
export async function updateProxyList(network: string, data: ProxyData, fileName?:string) {
  
  await updateList(network, "proxies", data);
}

export async function updateMetaList(network: string, data: MetaContractData, fileName?:string) {
  
  await updateList(network, "staticContracts", data)
}
export async function updateList(network: string, fieldName: string, data: object, fileName?:string){
  
  let factoryStateConfig = await readFactoryState();
  let output:Array<any> = factoryStateConfig[network][fieldName] === undefined ?  [] : factoryStateConfig[network][fieldName]   
  output.push(data)
  // write new data to config file
  await writeFactoryState(network, fieldName, output, fileName);
}
