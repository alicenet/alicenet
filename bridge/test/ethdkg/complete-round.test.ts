import { validators4 } from "./assets/4-validators-successful-case";
import { validators10 } from "./assets/10-validators-successful-case";
import { completeETHDKGRound } from "./setup";

describe("ETHDKG: Complete an ETHDKG Round", () => {
  it("completes happy path with 4 validators", async function () {
    await completeETHDKGRound(validators4);
  });

  it("completes happy path with 10 validators", async function () {
    await completeETHDKGRound(validators10);
  });
});
