import { BytesLike, ContractTransactionReceipt } from "ethers";

export class InitializerArgsError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "InitializerArgsError";
  }
}

export interface ArgData {
  [key: string]:
    | string
    | number
    | boolean
    | bigint
    | BytesLike
    | ArrayLike<string | number | boolean | bigint | BytesLike>;
}
export interface ContractArgs {
  [key: string]: ArgData;
}
export interface DeploymentArgs {
  constructor: ContractArgs;
  initializer: ContractArgs;
}

export interface ContractDescriptor {
  name: string;
  fullyQualifiedName: string;
  deployGroup: string;
  deployGroupIndex: number;
  deployType: string;
  constructorArgs: Array<any>;
  initializerArgs: Array<any>;
}

export interface DeploymentConfig {
  name: string;
  fullyQualifiedName: string;
  salt: string;
  deployGroup: string;
  deployGroupIndex: string;
  deployType: string;
  constructorArgs: ArgData;
  initializerArgs: ArgData;
}

export interface DeploymentConfigWrapper {
  [key: string]: DeploymentConfig;
}

export type DeploymentList = {
  [key: string]: Array<DeploymentConfig>;
};

export type FactoryData = {
  address: string;
  owner?: string;
  gas: bigint;
};

export type DeployCreateData = {
  name: string;
  address: string;
  factoryAddress: string;
  gas: bigint;
  constructorArgs?: any;
  receipt?: ContractTransactionReceipt;
};

export type MetaContractData = {
  metaAddress: string;
  salt: string;
  templateName: string;
  templateAddress: string;
  factoryAddress: string;
  gas: bigint;
  initCallData: string;
  receipt?: ContractTransactionReceipt;
};

export type TemplateData = {
  name: string;
  address: string;
  factoryAddress: string;
  gas: bigint;
  receipt?: ContractTransactionReceipt;
  constructorArgs?: string;
};

export type ProxyData = {
  proxyAddress: string;
  salt: BytesLike;
  logicName?: string;
  logicAddress?: string;
  factoryAddress: string;
  gas: bigint;
  receipt?: ContractTransactionReceipt;
  initCallData?: BytesLike;
};
