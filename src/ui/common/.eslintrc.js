// Source for this ESLint configuration:
// https://robertcooper.me/post/using-eslint-and-prettier-in-a-typescript-project
module.exports = {
  parser: "@typescript-eslint/parser",
  // Specifies the ESLint parser
  parserOptions: {
    ecmaVersion: 2020,
    // Allows for the parsing of modern ECMAScript features
    sourceType: "module",
    // Allows for the use of imports
    ecmaFeatures: {
      jsx: true // Allows for the parsing of JSX

    }
  },
  settings: {
    react: {
      version: "detect" // Tells eslint-plugin-react to automatically detect the version of React to use

    }
  },
  extends: ["plugin:react/recommended", "plugin:@typescript-eslint/recommended", "plugin:prettier/recommended"],
  plugins: ['simple-import-sort', "unused-imports", "react-hooks"],
  rules: {
    // Place to specify ESLint rules. Can be used to overwrite rules specified from the extended configs
    // e.g. "@typescript-eslint/explicit-function-return-type": "off",
    // Since we're using Typescript, checking prop-types is no longer needed. This line stops the prop-types errors from happening during lint.
    'react/prop-types': 0,
    'simple-import-sort/imports': 'error',
    'simple-import-sort/exports': 'error',
    'no-unused-vars': 'off',
    'unused-imports/no-unused-imports': 'error',
    'react/jsx-child-element-spacing': 'off',
    'react-hooks/rules-of-hooks': 'error', // Checks rules of Hooks
    'react-hooks/exhaustive-deps': 'warn', // Checks effect dependencies
  },
};
