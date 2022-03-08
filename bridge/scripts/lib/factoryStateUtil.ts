import { BigNumberish, BytesLike } from "ethers";
import fs, { write } from "fs";
import { DEPLOYMENT_CONFIG_PATH, FACTORY_STATE_CONFIG_PATH} from "./constants";
import toml from"@iarna/toml"
import { readDeploymentConfig, readFactoryStateData } from "./baseConfigUtil";
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



export async function getDefaultFactoryAddress(network:string): Promise<string> {
  //fetch whats in the factory config file
  let config = await readFactoryStateData();
  return config[network].defaultFactoryData.address;
}



async function writeFactoryConfig(network: string, fieldName: string, fieldData: any){
  let factoryStateConfig
  if(fs.existsSync(FACTORY_STATE_CONFIG_PATH)){
    factoryStateConfig = await readFactoryStateData();
    factoryStateConfig[network] = factoryStateConfig[network] === undefined ? {} : factoryStateConfig[network];
    factoryStateConfig[network][fieldName] = fieldData;
  }else{
    factoryStateConfig = {
      network: {
        [fieldName]: fieldData
      }
    }
  }
  let data = toml.stringify(factoryStateConfig)
  fs.writeFileSync(FACTORY_STATE_CONFIG_PATH, data);
}

export async function updateDefaultFactoryData(network: string, data: FactoryData) {
  await writeFactoryConfig(network, "defaultFactoryData", data)
  await writeFactoryConfig(network, "defaultFactoryAddress", data.address)
  let deploymentConfig:any = readDeploymentConfig();
  for(let i in deploymentConfig){
    //network
    for(let j in deploymentConfig[i]){
      //args
      for(let k = 0; k< deploymentConfig[i][j].length; k++){
        let name = Object.keys(deploymentConfig[i][j][k])[0]
        if(name === "selfAddr_" || name === "owner_" || name === "factoryAddr_"){
          if(name === "owner_" && data.owner !== undefined){
            deploymentConfig[i][j][k][name] = data.owner   
          } else {
            deploymentConfig[i][j][k][name] = data.address 
          }
        }
      }  
    }  
  }  
  let input = toml.stringify(deploymentConfig)
  fs.writeFileSync(DEPLOYMENT_CONFIG_PATH, input);
}

export async function updateDeployCreateList(network: string, data: DeployCreateData) {
  await updateList(network, "rawDeployments", data)
}

export async function updateTemplateList(network: string, data: TemplateData) {
  await updateList(network, "templates", data);
}

/**
 * @description pulls in the factory config data and adds proxy data
 * to the proxy array
 * @param data object that contains the proxies
 * logic contract name, address, and proxy address
 */
export async function updateProxyList(network: string, data: ProxyData) {
  await updateList(network, "proxies", data);
}

export async function updateMetaList(network: string, data: MetaContractData) {
  await updateList(network, "staticContracts", data)
}
export async function updateList(network: string, fieldName: string, data: object){
  let factoryStateConfig = await readFactoryStateData();
  let output:Array<any> = factoryStateConfig[network][fieldName] === undefined ?  [] : factoryStateConfig[network][fieldName]   
  output.push(data)
  // write new data to config file
  await writeFactoryConfig(network, fieldName, output);
}
