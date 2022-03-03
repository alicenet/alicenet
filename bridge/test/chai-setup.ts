import chai from "chai";
import chaiAsPromised from "chai-as-promised";
import { chaiEthers } from "chai-ethers";
import { solidity } from "ethereum-waffle";
chai.use(solidity);
chai.use(chaiEthers);
chai.use(chaiAsPromised);
chai.config.includeStack = true;
export = chai;
