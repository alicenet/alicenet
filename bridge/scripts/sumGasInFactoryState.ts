import fs from "fs";
import { FactoryData } from "./lib/deployment/factoryStateUtil";

async function main() {
  // read the file
  const rawData = fs.readFileSync("./deployments/testnet/factoryState.json");
  const State = await JSON.parse(rawData.toString("utf8"));
  let gas = 0;
  const defaultFactory: FactoryData = State.defaultFactoryData;
  if (defaultFactory.gas !== undefined) {
    gas += defaultFactory.gas;
  }
  for (const contract of State.templates) {
    gas += contract.gas;
  }
  for (const contract of State.staticContracts) {
    gas += contract.gas;
  }
  for (const contract of State.deployCreates) {
    gas += contract.gas;
  }
  for (const contract of State.proxies) {
    gas += contract.gas;
  }
  console.log(gas);
}
main()
  .then(() => {
    return 0;
  })
  .catch((error) => {
    console.error(error);
    throw new Error("unexpected error");
  });
