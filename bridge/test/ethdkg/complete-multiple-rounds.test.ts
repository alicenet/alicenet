import { validators4 } from "./assets/4-validators-successful-case";
import { validators10 } from "./assets/10-validators-successful-case";
import { completeETHDKGRound, expect, registerValidators } from "./setup";


describe("ETHDKG: Complete an ETHDKG Round and change validators", () => {
  it("completes ETHDKG with 10 validators then change to 4 validators", async function () {
    let [ethdkg, validatorPool, expectedNonce, ] = await completeETHDKGRound(
      validators10
    );
    expect(expectedNonce).eq(1);
    await validatorPool.unregisterAllValidators();
    [, , expectedNonce] = await completeETHDKGRound(validators4, {ethdkg, validatorPool});
    expect(expectedNonce).eq(2);
  });

  it("completes ETHDKG with 10 validators then a validator try to register without registration open", async function () {
    let [ethdkg, validatorPool, expectedNonce ,] = await completeETHDKGRound(
      validators10
    );

    await expect(registerValidators(ethdkg, validatorPool, validators10, expectedNonce))
    .to.be.revertedWith("ETHDKG: Cannot register at the moment")
  });
});