import { expect } from "chai";
import { run } from "hardhat";
import { MOCK } from "../../scripts/lib/constants";
import { getDefaultFactoryAddress } from "../../scripts/lib/factoryStateUtils";
import {
  getAccounts,
  predictFactoryAddress
} from "./Setup";

describe("Cli tasks", async () => {
  let firstOwner: string;
  let firstDelegator: string;
  let accounts: Array<string> = [];

  beforeEach(async () => {
    accounts = await getAccounts();
    process.env.silencer = "true";
    //set owner and delegator
    firstOwner = accounts[0];
    firstDelegator = accounts[1];
  });

  xit("deploy factory with cli", async () => {
    let futureFactoryAddress = await predictFactoryAddress(firstOwner);
    let factoryAddress = await run("deployFactory");
    //check if the address is the predicted
    expect(factoryAddress).to.equal(futureFactoryAddress);
    let defaultFactoryAddress = await getDefaultFactoryAddress();
    expect(defaultFactoryAddress).to.equal(factoryAddress);
  });


  xit("deploy mock with deploystatic", async () => {
    await run("deployMetamorphic", {
      contractName: MOCK,
      constructorArgs: ["2", "s"]
    });
  });
});
