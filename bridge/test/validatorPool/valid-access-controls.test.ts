import { expect } from "../chai-setup";
import { factoryCallAny, Fixture, getFixture } from "../setup";
import { commitSnapshots } from "./setup";

describe("ValidatorPool Access Control: An user with admin role should be able to:", async function () {
  let fixture: Fixture;
  const maxNumValidators = 5;
  const stakeAmount = 20000;

  beforeEach(async function () {
    fixture = await getFixture(false, true);
  });

  it("Set a minimum stake", async function () {
    const rcpt = await factoryCallAny(
      fixture,
      "validatorPool",
      "setStakeAmount",
      [stakeAmount]
    );
    expect(rcpt.status).to.equal(1);
  });

  it("Set a maximum number of validators", async function () {
    const rcpt = await factoryCallAny(
      fixture,
      "validatorPool",
      "setMaxNumValidators",
      [maxNumValidators]
    );
    expect(rcpt.status).to.equal(1);
  });

  it("Schedule maintenance", async function () {
    const rcpt = await factoryCallAny(
      fixture,
      "validatorPool",
      "scheduleMaintenance"
    );
    expect(rcpt.status).to.equal(1);
  });

  it("Register validators", async function () {
    await expect(
      factoryCallAny(fixture, "validatorPool", "registerValidators", [
        ["0x000000000000000000000000000000000000dEaD"],
        [1],
      ])
    ).to.be.revertedWith("ERC721: owner query for nonexistent token");
  });

  it("Initialize ETHDKG", async function () {
    await expect(
      factoryCallAny(fixture, "validatorPool", "initializeETHDKG")
    ).to.be.revertedWith(
      "ETHDKG: Minimum number of validators staked not met!"
    );
  });

  it("Unregister validators", async function () {
    await expect(
      factoryCallAny(fixture, "validatorPool", "unregisterValidators", [
        ["0x000000000000000000000000000000000000dEaD"],
      ])
    ).to.be.revertedWith(
      "ValidatorPool: There are not enough validators to be removed!"
    );
  });

  it("Pause consensus", async function () {
    await commitSnapshots(fixture, 1);
    await expect(
      factoryCallAny(
        fixture,
        "validatorPool",
        "pauseConsensusOnArbitraryHeight",
        [1]
      )
    ).to.be.revertedWith("ValidatorPool: Condition not met to stop consensus!");
  });
});
