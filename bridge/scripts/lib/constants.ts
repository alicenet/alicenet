import "process";
export const MULTICALL_GAS_LIMIT = "10000000";
export const DEPLOYMENT_LIST_FPATH = "/deploymentList";
export const DEPLOYMENT_ARGS_TEMPLATE_FPATH = "/deploymentArgsTemplate";
export const DEFAULT_CONFIG_OUTPUT_DIR = "../scripts/base-files";
export const BASE_CONFIG_PATH = `../scripts/base-files/baseConfig`;
export const CONTRACT_ADDR = "contractAddr";
export const DEPLOY_ALL_CONTRACTS = "deployAllContracts";
export const TASK_DEPLOY_ALL_CONTRACTS = "deploy-all-contracts";
export const DEPLOY_CREATE = "deployCreate";
export const TASK_DEPLOY_CREATE = "deploy-create";
export const DEPLOY_CREATE2 = "deployCreate2";
export const TASK_DEPLOY_CREATE2 = "deploy-create2";
export const DEPLOY_FACTORY = "deployFactory";
export const TASK_DEPLOY_FACTORY = "deploy-factory";
export const DEPLOY_METAMORPHIC = "deployMetamorphic";
export const TASK_DEPLOY_METAMORPHIC = "deploy-metamorphic";
export const DEPLOY_PROXY = "deployProxy";
export const TASK_DEPLOY_PROXY = "deploy-proxy";
export const DEPLOY_STATIC = "deployStatic";
export const TASK_DEPLOY_STATIC = "deploy-static";
export const DEPLOY_TEMPLATE = "deployTemplate";
export const TASK_DEPLOY_TEMPLATE = "deploy-template";
export const TASK_DEPLOY_CONTRACTS = "deploy-contracts";
export const DEPLOY_UPGRADEABLE_PROXY = "deployUpgradeableProxy";
export const TASK_DEPLOY_UPGRADEABLE_PROXY = "deploy-upgradeable-proxy";
export const DEPLOYED_PROXY = "DeployedProxy";
export const DEPLOYED_RAW = "DeployedRaw";
export const DEPLOYED_STATIC = "DeployedStatic";
export const DEPLOYED_TEMPLATE = "DeployedTemplate";
export const DEPLOYMENT_ARG_PATH = `../scripts/base-files/deploymentArgs`;
export const DEPLOYMENT_LIST_PATH = `../scripts/base-files/deploymentList`;
export const END_POINT = "MockEndPoint";
export const FACTORY_DEPLOYMENT: string = "factory";
export const FACTORY_STATE_PATH = `../scripts/generated/factoryState`;
export const LOGIC_ADDR = "LogicAddress";
export const ALICENET_FACTORY = "AliceNetFactory";
export const META_ADDR = "MetaAddress";
export const MOCK = "MockBaseContract";
export const MOCK_FACTORY = "MockFactory";
export const MOCK_INITIALIZABLE = "MockInitializable";
export const MULTI_CALL_DEPLOY_PROXY = "multiCallDeployProxy";
export const TASK_MULTI_CALL_DEPLOY_PROXY = "multi-call-deploy-proxy";
export const TASK_MULTI_CALL_DEPLOY_METAMORPHIC =
  "multi-call-deploy-metamorphic";
export const TASK_FULL_MULTI_CALL_DEPLOY_PROXY = "full-multi-call-deploy-proxy";
export const MULTI_CALL_UPGRADE_PROXY = "multiCallUpgradeProxy";
export const TASK_MULTI_CALL_UPGRADE_PROXY = "multi-call-upgrade-proxy";
export const ONLY_PROXY = "onlyProxy";
export const PROXY = "Proxy";
export const PROXY_ADDR = "ProxyAddress";
export const RECEIPT = "receipt";
export const STATIC_DEPLOYMENT: string = "deployStatic";
export const TEMPLATE_ADDR = "TemplateAddress";
export const UPGRADE_DEPLOYED_PROXY = "upgradeDeployedProxy";
export const TASK_UPGRADE_DEPLOYED_PROXY = "upgrade-deployed-proxy";
export const UPGRADE_PROXY = "upgradeProxy";
export const TASK_UPGRADE_PROXY = "upgrade-proxy";
export const UPGRADEABLE_DEPLOYMENT: string = "deployUpgradeable";
export const TASK_UPGRADEABLE_DEPLOYMENT: string = "deploy-upgradeable";
export const UTILS = "Utils";
export const INITIALIZER = "initialize";
export const HARDHAT_CHAIN_ID = 1337;
export const env = (): string => {
  let _env = process.env.DEPLOYMENT_ENVIRONMENT;
  if (typeof _env === "undefined") {
    _env = "dev";
  }
  return _env;
};
