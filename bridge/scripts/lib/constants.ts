import "process";
export const MULTICALL_GAS_LIMIT = "10000000";
// contract names
export const ALICENET_FACTORY = "AliceNetFactory";
export const END_POINT = "MockEndPoint";
export const LOCK_UP = "Lockup";
export const MOCK_FACTORY = "MockFactory";
export const MOCK = "MockBaseContract";
export const STAKING_ROUTER_V1 = "StakingRouterV1";
export const UTILS = "Utils";
// function names
export const FUNCTION_INITIALIZE = "initialize";
export const FUNCTION_DEPLOY_CREATE2 = "deployCreate2";
export const FUNCTION_DEPLOY_CREATE = "deployCreate";
export const FUNCTION_DEPLOY_METAMORPHIC = "deployMetamorphic";
export const FUNCTION_UPGRADE_PROXY = "upgradeProxy";
// Factory event names
export const EVENT_DEPLOYED_PROXY = "DeployedProxy";
export const EVENT_DEPLOYED_RAW = "DeployedRaw";
export const EVENT_DEPLOYED_TEMPLATE = "DeployedTemplate";
export const PROXY = "Proxy";
// Event variable names
export const CONTRACT_ADDR = "contractAddr";

// Hardhat CLI Task names
export const DEPLOYMENT_ARG_PATH = `../scripts/base-files/deploymentArgs`;
export const DEPLOYMENT_LIST_PATH = `../scripts/base-files/deploymentList`;
export const FACTORY_DEPLOYMENT: string = "factory";
export const FACTORY_STATE_PATH = `../scripts/generated/factoryState`;
export const LOGIC_ADDR = "LogicAddress";
export const META_ADDR = "MetaAddress";
export const MOCK_INITIALIZABLE = "MockInitializable";
export const ONLY_PROXY = "onlyProxy";
export const PROXY_ADDR = "ProxyAddress";
export const RECEIPT = "receipt";
export const UPGRADEABLE_DEPLOYMENT: string = "deployUpgradeable";

// default paths
export const DEPLOYMENT_LIST_FPATH = "/deploymentList";
export const DEPLOYMENT_ARGS_TEMPLATE_FPATH = "/deploymentArgsTemplate";
export const DEFAULT_CONFIG_DIR = "../scripts/base-files";
export const DEFAULT_FACTORY_STATE_OUTPUT_DIR = "../scripts/generated";
export const BASE_CONFIG_PATH = `../scripts/base-files/baseConfig`;
export const HARDHAT_CHAIN_ID = 1337;
export const env = (): string => {
  let _env = process.env.DEPLOYMENT_ENVIRONMENT;
  if (typeof _env === "undefined") {
    _env = "dev";
  }
  return _env;
};
