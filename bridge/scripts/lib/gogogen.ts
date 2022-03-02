
import { task } from "hardhat/config";
import { HardhatRuntimeEnvironment, } from "hardhat/types";
import fs from 'fs';

task("go-go-gen", "Builds generate.go file from all json files in path")
.addParam("in", "relative path of the output files")
.addParam("out", "relative path of the output files")
.addParam("pkg", "pkg the go generate command should use")
.setAction(async (taskArgs, hre: HardhatRuntimeEnvironment) => {
    let outputData   = `package bridge\n\n//go:generate mkdir -p ${taskArgs.out}\n`;
    fs.readdirSync(taskArgs.in).forEach(file => {
        let source = file.replace(".json", "");
        outputData = outputData + `//go:generate abigen --abi ${taskArgs.in}/${source}.json --pkg ${taskArgs.pkg} --type ${source} --out ${taskArgs.out}/${source}.go\n`;
    });
    fs.writeFileSync("generate.go", outputData);
});