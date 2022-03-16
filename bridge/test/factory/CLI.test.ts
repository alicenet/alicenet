import { expect } from "chai";
import { run } from "hardhat";
import { MOCK } from "../../scripts/lib/constants";
import { getDefaultFactoryAddress } from "../../scripts/lib/factoryStateUtils";
import { getAccounts, predictFactoryAddress } from "./Setup";

describe("Cli tasks", async () => {
  let firstOwner: string;
  let accounts: Array<string> = [];

  beforeEach(async () => {
    accounts = await getAccounts();
    process.env.silencer = "true";
    // set owner and delegator
    firstOwner = accounts[0];
  });

  xit("deploy factory with cli", async () => {
    const futureFactoryAddress = await predictFactoryAddress(firstOwner);
    const factoryAddress = await run("deployFactory");
    // check if the address is the predicted
    expect(factoryAddress).to.equal(futureFactoryAddress);
    const defaultFactoryAddress = await getDefaultFactoryAddress();
    expect(defaultFactoryAddress).to.equal(factoryAddress);
  });

  xit("deploy mock with deploystatic", async () => {
    await run("deployMetamorphic", {
      contractName: MOCK,
      constructorArgs: ["2", "s"],
    });
  });
});
