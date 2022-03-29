import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../../chai-setup";
import {
  callFunctionAndGetReturnValues,
  factoryCallAny,
  Fixture,
  getFixture,
} from "../../setup";
import { getState, init, showState } from "./setup";

describe("Testing MadByte Distribution methods", async () => {
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let user2: SignerWithAddress;
  let fixture: Fixture;
  let minMadBytes = 0;
  let eth = 4;
  let ethIn: BigNumber;

  beforeEach(async function () {
    fixture = await getFixture();
    let signers = await ethers.getSigners();
    [admin, user, user2] = signers;
    await init(fixture);
    showState("Initial", await getState(fixture));
    await factoryCallAny(fixture, "madByte", "setAdmin", [admin.address]);
    ethIn = ethers.utils.parseEther(eth.toString());
  });

  it("Should correctly distribute", async () => {
    const [madBytes] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "mint",
      admin,
      [minMadBytes],
      ethIn
    );
    const splits = [250, 250, 250, 250];
    await fixture.madByte.setSplits(splits[0], splits[1], splits[2], splits[3]);
    const [distribution] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "distribute",
      admin,
      []
    );
    let distributable = ethIn.sub(await fixture.madByte.getPoolBalance());
    expect(
      (await ethers.provider.getBalance(fixture.validatorStaking.address)).mul(
        1000
      )
    ).to.be.equal(distributable.mul(splits[0]));
    expect(
      (await ethers.provider.getBalance(fixture.publicStaking.address)).mul(
        1000
      )
    ).to.be.equal(distributable.mul(splits[1]));
    expect(
      (
        await ethers.provider.getBalance(
          fixture.liquidityProviderStaking.address
        )
      ).mul(1000)
    ).to.be.equal(distributable.mul(splits[2]));
    expect(
      (await ethers.provider.getBalance(fixture.foundation.address)).mul(1000)
    ).to.be.equal(distributable.mul(splits[3]));
  });

  it("Should correctly distribute big amount of eth", async () => {
    ethIn = ethers.utils.parseEther("70000000000".toString());
    const splits = [250, 250, 250, 250];
    await fixture.madByte.setSplits(splits[0], splits[1], splits[2], splits[3]);
    // Burn previous supply
    const [madBytes] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "mint",
      admin,
      [minMadBytes],
      ethIn
    );
    const [distribution] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "distribute",
      admin,
      []
    );
    let distributable = ethIn.sub(await fixture.madByte.getPoolBalance());
    expect(
      (await ethers.provider.getBalance(fixture.validatorStaking.address)).mul(
        1000
      )
    ).to.be.equal(distributable.mul(splits[0]));
    expect(
      (await ethers.provider.getBalance(fixture.publicStaking.address)).mul(
        1000
      )
    ).to.be.equal(distributable.mul(splits[1]));
    expect(
      (
        await ethers.provider.getBalance(
          fixture.liquidityProviderStaking.address
        )
      ).mul(1000)
    ).to.be.equal(distributable.mul(splits[2]));
    expect(
      (await ethers.provider.getBalance(fixture.foundation.address)).mul(1000)
    ).to.be.equal(distributable.mul(splits[3]));
  });

  it("Should distribute without foundation", async () => {
    const [madBytes] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "mint",
      admin,
      [minMadBytes],
      ethIn
    );
    const splits = [350, 350, 300, 0];
    await fixture.madByte.setSplits(splits[0], splits[1], splits[2], splits[3]);
    const [distribution] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "distribute",
      admin,
      []
    );
    let distributable = ethIn.sub(await fixture.madByte.getPoolBalance());
    expect(
      (await ethers.provider.getBalance(fixture.validatorStaking.address)).mul(
        1000
      )
    ).to.be.equal(distributable.mul(splits[0]));
    expect(
      (await ethers.provider.getBalance(fixture.publicStaking.address)).mul(
        1000
      )
    ).to.be.equal(distributable.mul(splits[1]));
    expect(
      (
        await ethers.provider.getBalance(
          fixture.liquidityProviderStaking.address
        )
      ).mul(1000)
    ).to.be.equal(distributable.mul(splits[2]));
    expect(
      (await ethers.provider.getBalance(fixture.foundation.address)).mul(1000)
    ).to.be.equal(distributable.mul(splits[3]));
  });

  it("Should distribute without liquidityProviderStaking", async () => {
    const [madBytes] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "mint",
      admin,
      [minMadBytes],
      ethIn
    );
    const splits = [350, 350, 0, 300];
    await fixture.madByte.setSplits(splits[0], splits[1], splits[2], splits[3]);
    const [distribution] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "distribute",
      admin,
      []
    );
    let distributable = ethIn.sub(await fixture.madByte.getPoolBalance());
    expect(
      (await ethers.provider.getBalance(fixture.validatorStaking.address)).mul(
        1000
      )
    ).to.be.equal(distributable.mul(splits[0]));
    expect(
      (await ethers.provider.getBalance(fixture.publicStaking.address)).mul(
        1000
      )
    ).to.be.equal(distributable.mul(splits[1]));
    expect(
      (
        await ethers.provider.getBalance(
          fixture.liquidityProviderStaking.address
        )
      ).mul(1000)
    ).to.be.equal(distributable.mul(splits[2]));
    expect(
      (await ethers.provider.getBalance(fixture.foundation.address)).mul(1000)
    ).to.be.equal(distributable.mul(splits[3]));
  });

  it("Should distribute without publicStaking", async () => {
    const [madBytes] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "mint",
      admin,
      [minMadBytes],
      ethIn
    );
    const splits = [350, 350, 0, 300];
    await fixture.madByte.setSplits(splits[0], splits[1], splits[2], splits[3]);
    const [distribution] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "distribute",
      admin,
      []
    );
    let distributable = ethIn.sub(await fixture.madByte.getPoolBalance());
    expect(
      (await ethers.provider.getBalance(fixture.validatorStaking.address)).mul(
        1000
      )
    ).to.be.equal(distributable.mul(splits[0]));
    expect(
      (await ethers.provider.getBalance(fixture.publicStaking.address)).mul(
        1000
      )
    ).to.be.equal(distributable.mul(splits[1]));
    expect(
      (
        await ethers.provider.getBalance(
          fixture.liquidityProviderStaking.address
        )
      ).mul(1000)
    ).to.be.equal(distributable.mul(splits[2]));
    expect(
      (await ethers.provider.getBalance(fixture.foundation.address)).mul(1000)
    ).to.be.equal(distributable.mul(splits[3]));
  });

  it("Should distribute without validatorStaking", async () => {
    const [madBytes] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "mint",
      admin,
      [minMadBytes],
      ethIn
    );
    const splits = [0, 350, 350, 300];
    await fixture.madByte.setSplits(splits[0], splits[1], splits[2], splits[3]);
    const [distribution] = await callFunctionAndGetReturnValues(
      fixture.madByte,
      "distribute",
      admin,
      []
    );
    let distributable = ethIn.sub(await fixture.madByte.getPoolBalance());
    expect(
      (await ethers.provider.getBalance(fixture.validatorStaking.address)).mul(
        1000
      )
    ).to.be.equal(distributable.mul(splits[0]));
    expect(
      (await ethers.provider.getBalance(fixture.publicStaking.address)).mul(
        1000
      )
    ).to.be.equal(distributable.mul(splits[1]));
    expect(
      (
        await ethers.provider.getBalance(
          fixture.liquidityProviderStaking.address
        )
      ).mul(1000)
    ).to.be.equal(distributable.mul(splits[2]));
    expect(
      (await ethers.provider.getBalance(fixture.foundation.address)).mul(1000)
    ).to.be.equal(distributable.mul(splits[3]));
  });
});
