import { Contract } from "ethers";
import { assertConstantReturnsCorrectErrorCode, deployLibrary } from "./setup";

describe("Testing MagicTokenTransfer error codes", async () => {
  let magicTokenTransferErrorCodesContract: Contract;

  beforeEach(async function () {
    magicTokenTransferErrorCodesContract = await deployLibrary(
      "MagicTokenTransferErrorCodes"
    );
  });

  it("MAGICTOKENTRANSFER_TRANSFER_FAILED returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      magicTokenTransferErrorCodesContract.MAGICTOKENTRANSFER_TRANSFER_FAILED,
      2100
    );
  });
});
