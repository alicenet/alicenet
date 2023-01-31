module.exports = {
  configureYulOptimizer: true,
  skipFiles: [
    "test",
    "mocks",
    "node_modules",
    "test_sol",
    "lib",
    "./contracts/BonusPool.sol",
    "./contracts/Lockup.sol",
    "./contracts/RewardPool.sol",
  ],
  mocha: {
    grep: "@skip-on-coverage", // Find everything with this tag
    invert: true, // Run the grep's inverse set.
  },
  //   solcOptimizerDetails: {
  //     peephole: false,
  //     inliner: false,
  //     jumpdestRemover: false,
  //     orderLiterals: true, // <-- TRUE! Stack too deep when false
  //     deduplicate: false,
  //     cse: false,
  //     constantOptimizer: false,
  //     yul: true,
  //   },
};
