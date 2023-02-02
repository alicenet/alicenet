module.exports = {
  configureYulOptimizer: true,
  skipFiles: [
    "test",
    "mocks",
    "node_modules",
    "test_sol",
    "BonusPool.sol",
    "Lockup.sol",
    "RewardPool.sol",
    "StakingRouterV1.sol",
    "utils/auth",
    "libraries/errors",
    "interfaces/",
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
