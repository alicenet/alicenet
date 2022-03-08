import "process";

//deployment Types
export const FACTORY_STATE_CONFIG_PATH = `../scripts/base-files/factoryStateConfig`;
export const DEPLOYMENT_CONFIG_PATH = `../scripts/base-files/deploymentConfig`;
export const BASE_CONFIG_PATH = `../scripts/base-files/baseConfig`;
export const STATIC_DEPLOYMENT: string = "deployStatic";
export const UPGRADEABLE_DEPLOYMENT: string = "deployUpgradeable";
export const FACTORY_DEPLOYMENT: string = "factory";
export const MOCK_INITIALIZABLE = "MockInitializable";
export const MOCK_FACTORY = "MockFactory";
export const LOGIC_ADDR = "LogicAddress";
export const PROXY_ADDR = "ProxyAddress";
export const UTILS = "Utils";
export const META_ADDR = "MetaAddress";
export const TEMPLATE_ADDR = "TemplateAddress";
export const END_POINT = "MockEndPoint";
export const DEPLOYED_STATIC = "DeployedStatic";
export const DEPLOYED_PROXY = "DeployedProxy";
export const DEPLOYED_RAW = "DeployedRaw";
export const DEPLOYED_TEMPLATE = "DeployedTemplate";
export const CONTRACT_ADDR = "contractAddr";
export const MADNET_FACTORY = "MadnetFactory";
export const MOCK = "MockBaseContract";
export const RECEIPT = "receipt";
export const DEPLOY_CREATE = "deployCreate";
export const DEPLOY_CREATE2 = "deployCreate2";
export const DEPLOY_PROXY = "deployProxy";
export const DEPLOY_TEMPLATE = "deployTemplate";
export const DEPLOY_STATIC = "deployStatic";
export const UPGRADE_PROXY = "upgradeProxy";
export const PROXY = "Proxy";
export const env = (): string => {
  let _env = process.env["DEPLOYMENT_ENVIRONMENT"];
  if (typeof _env === "undefined") {
    _env = "dev";
  }
  return _env;
};
