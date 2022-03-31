import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import {
  callFunctionAndGetReturnValues,
  factoryCallAnyFixture,
  Fixture,
  getFixture,
} from "../setup";
import { getState, showState } from "./setup";

describe("Testing BToken Distribution methods", async () => {
  let admin: SignerWithAddress;
  let fixture: Fixture;
  const minBTokens = 0;
  const eth = 4;
  let ethIn: BigNumber;

  beforeEach(async function () {
    fixture = await getFixture();
    const signers = await ethers.getSigners();
    [admin] = signers;
    showState("Initial", await getState(fixture));
    await factoryCallAnyFixture(fixture, "bToken", "setAdmin", [admin.address]);
    ethIn = ethers.utils.parseEther(eth.toString());
  });

  it("Should correctly distribute", async () => {
    await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mint",
      admin,
      [minBTokens],
      ethIn
    );
    const splits = [250, 250, 250, 250];
    await fixture.bToken.setSplits(splits[0], splits[1], splits[2], splits[3]);
    await callFunctionAndGetReturnValues(
      fixture.bToken,
      "distribute",
      admin,
      []
    );
    const distributable = ethIn.sub(await fixture.bToken.getPoolBalance());
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
    await fixture.bToken.setSplits(splits[0], splits[1], splits[2], splits[3]);
    // Burn previous supply
    await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mint",
      admin,
      [minBTokens],
      ethIn
    );
    await callFunctionAndGetReturnValues(
      fixture.bToken,
      "distribute",
      admin,
      []
    );
    const distributable = ethIn.sub(await fixture.bToken.getPoolBalance());
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
    await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mint",
      admin,
      [minBTokens],
      ethIn
    );
    const splits = [350, 350, 300, 0];
    await fixture.bToken.setSplits(splits[0], splits[1], splits[2], splits[3]);
    await callFunctionAndGetReturnValues(
      fixture.bToken,
      "distribute",
      admin,
      []
    );
    const distributable = ethIn.sub(await fixture.bToken.getPoolBalance());
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
    await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mint",
      admin,
      [minBTokens],
      ethIn
    );
    const splits = [350, 350, 0, 300];
    await fixture.bToken.setSplits(splits[0], splits[1], splits[2], splits[3]);
    await callFunctionAndGetReturnValues(
      fixture.bToken,
      "distribute",
      admin,
      []
    );
    const distributable = ethIn.sub(await fixture.bToken.getPoolBalance());
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
    await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mint",
      admin,
      [minBTokens],
      ethIn
    );
    const splits = [350, 350, 0, 300];
    await fixture.bToken.setSplits(splits[0], splits[1], splits[2], splits[3]);
    await callFunctionAndGetReturnValues(
      fixture.bToken,
      "distribute",
      admin,
      []
    );
    const distributable = ethIn.sub(await fixture.bToken.getPoolBalance());
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
    await callFunctionAndGetReturnValues(
      fixture.bToken,
      "mint",
      admin,
      [minBTokens],
      ethIn
    );
    const splits = [0, 350, 350, 300];
    await fixture.bToken.setSplits(splits[0], splits[1], splits[2], splits[3]);
    await callFunctionAndGetReturnValues(
      fixture.bToken,
      "distribute",
      admin,
      []
    );
    const distributable = ethIn.sub(await fixture.bToken.getPoolBalance());
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
