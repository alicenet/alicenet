module.exports = {
  configureYulOptimizer: true,
  solcOptimizerDetails: {
    peephole: false,
    inliner: true,
    jumpdestRemover: true,
    orderLiterals: true, // <-- TRUE! Stack too deep when false
    deduplicate: true,
    cse: false,
    constantOptimizer: false,
    yul: false,
  },
};
