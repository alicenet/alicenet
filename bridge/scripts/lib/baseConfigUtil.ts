import toml from "@iarna/toml";
import fs from "fs";
import { ethers } from "hardhat";

export const BASE_CONFIG_PATH = `../scripts/base-files/baseConfig`;

export function readBaseConfig() {
    if (fs.existsSync(BASE_CONFIG_PATH)) {
      let output = fs.readFileSync(BASE_CONFIG_PATH);
      let config = toml.parse(output.toString());
      return config;
    } else {
      throw new Error(`failed to get baseConfig`);
    }
}