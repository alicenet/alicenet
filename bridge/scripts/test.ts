import fs from "fs"
import toml from "@iarna/toml"


async function main() {
  let i = {
    field: 1
  }
  fs.appendFileSync("foo", "#bar")
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });

