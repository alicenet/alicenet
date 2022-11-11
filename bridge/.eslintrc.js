module.exports = {
  env: {
    browser: false,
    es2022: true,
    mocha: true,
    node: true,
  },
  plugins: ["@typescript-eslint", "import"],
  extends: [
    "standard",
    "plugin:node/recommended",
    "plugin:prettier/recommended",
    "plugin:import/recommended",
    "plugin:import/typescript",
  ],
  parser: "@typescript-eslint/parser",
  parserOptions: {
    ecmaVersion: 13,
    project: "./tsconfig.json",
  },
  settings: {
    "import/extensions": [".js", ".jsx", ".ts", ".tsx"],
    "import/parsers": {
      "@typescript-eslint/parser": [".ts", ".tsx"],
    },
    "import/resolver": {
      node: {
        extensions: [".js", ".jsx", ".ts", ".tsx"],
      },
    },
  },
  rules: {
    "no-unused-vars": [
      "error",
      { vars: "all", args: "after-used", ignoreRestSiblings: false },
    ],
    "node/no-unsupported-features/es-syntax": [
      "error",
      { ignores: ["modules"] },
    ],
    "@typescript-eslint/no-floating-promises": ["error"],
    "node/no-missing-import": [
      "error",
      {
        allowModules: [],
        resolvePaths: ["./test"],
        tryExtensions: [".js", ".json", ".node", ".ts", ".tsx"],
      },
    ],
    "node/no-unpublished-import": ["off"],
    "node/no-unpublished-require": ["off"],
  },
};
