import toml from "@iarna/toml";
import { BigNumber, BigNumberish, BytesLike, ContractReceipt } from "ethers";
import fs from "fs";
import { FACTORY_STATE_PATH } from "../constants";
import { readFactoryState } from "./deploymentConfigUtil";

export type FactoryData = {
  address: string;
  owner?: string;
  gas: BigNumber;
};

export type DeployCreateData = {
  name: string;
  address: string;
  factoryAddress: string;
  gas: BigNumber;
  constructorArgs?: any;
  receipt?: ContractReceipt;
};
export type MetaContractData = {
  metaAddress: string;
  salt: string;
  templateName: string;
  templateAddress: string;
  factoryAddress: string;
  gas: BigNumber;
  initCallData: string;
  receipt?: ContractReceipt;
};
export type TemplateData = {
  name: string;
  address: string;
  factoryAddress: string;
  gas: BigNumber;
  receipt?: ContractReceipt;
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
  gas: BigNumberish;
  receipt?: ContractReceipt;
  initCallData?: BytesLike;
};

export async function getDefaultFactoryAddress(
  network: string,
  usrPath?: string
): Promise<string> {
  const path =
    usrPath === undefined
      ? FACTORY_STATE_PATH
      : usrPath.replace(/\/+$/, "") + "/factoryState";
  // fetch whats in the factory config file
  const config = await readFactoryState(path);
  return config[network].defaultFactoryAddress;
}

async function writeFactoryState(
  network: string,
  fieldName: string,
  fieldData: any,
  usrPath?: string
) {
  if (process.env.silencer === undefined || process.env.silencer === "false") {
    const filePath =
      usrPath === undefined
        ? FACTORY_STATE_PATH
        : usrPath.replace(/\/+$/, "") + "/factoryState";
    let factoryStateConfig;
    if (fs.existsSync(filePath)) {
      factoryStateConfig = await readFactoryState(usrPath);
      factoryStateConfig[network] =
        factoryStateConfig[network] === undefined
          ? {}
          : factoryStateConfig[network];
      factoryStateConfig[network][fieldName] = fieldData;
    } else {
      factoryStateConfig = {
        [network]: {
          [fieldName]: fieldData,
        },
      };
    }
    const data = toml.stringify(factoryStateConfig);
    fs.writeFileSync(filePath, data);
  }
}

export async function updateDefaultFactoryData(
  network: string,
  data: FactoryData,
  usrPath?: string
) {
  await writeFactoryState(network, "defaultFactoryData", data, usrPath);
  await writeFactoryState(
    network,
    "defaultFactoryAddress",
    data.address,
    usrPath
  );
}

export async function updateDeployCreateList(
  network: string,
  data: DeployCreateData,
  usrPath?: string
) {
  await updateList(network, "rawDeployments", data, usrPath);
}

export async function updateTemplateList(
  network: string,
  data: TemplateData,
  usrPath?: string
) {
  await updateList(network, "templates", data, usrPath);
}

export async function updateProxyList(
  network: string,
  data: ProxyData,
  usrPath?: string
) {
  await updateList(network, "proxies", data, usrPath);
}

export async function updateMetaList(
  network: string,
  data: MetaContractData,
  usrPath?: string
) {
  await updateList(network, "staticContracts", data, usrPath);
}
export async function updateList(
  network: string,
  fieldName: string,
  data: any,
  usrPath?: string
) {
  if (process.env.silencer === undefined || process.env.silencer === "false") {
    const factoryStateConfig = await readFactoryState(usrPath);
    const output: Array<any> =
      factoryStateConfig[network][fieldName] === undefined
        ? []
        : factoryStateConfig[network][fieldName];
    output.push(data);
    if (data.receipt !== undefined) {
      data.receipt = undefined;
    }
    // write new state to config file
    await writeFactoryState(network, fieldName, output, usrPath);
  }
}

export async function getATokenMinterAddress(network: string) {
  // fetch whats in the factory config file
  const config = await readFactoryState(FACTORY_STATE_PATH);
  let proxies = config[network].proxies;
  for (let i = 0; i < proxies.length; i++) {
    let name = proxies[i].logicName;
    if (name === "ATokenMinter") {
      return proxies[i].proxyAddress;
    }
  }
}

export async function getBTokenAddress(network: string) {
  const config = await readFactoryState(FACTORY_STATE_PATH);
  let staticContracts = config[network].staticContracts;
  for (let i = 0; i < staticContracts.length; i++) {
    let name = staticContracts[i].templateName;
    if (name === "BToken") {
      return staticContracts[i].metaAddress;
    }
  }
}

export async function getATokenAddress(network: string) {
  const config = await readFactoryState(FACTORY_STATE_PATH);
  let staticContracts = config[network].staticContracts;
  for (let i = 0; i < staticContracts.length; i++) {
    let name = staticContracts[i].templateName;
    if (name === "AToken") {
      return staticContracts[i].metaAddress;
    }
  }
}
