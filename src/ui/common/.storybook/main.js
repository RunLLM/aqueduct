module.exports = {
  module: {
    rules: [{
      test: /\.(stories|story)\.mdx$/,
      use: [{
        loader: require.resolve('@storybook/mdx2-csf/loader'),
        options: {}
      }]
    }]
  },
  stories: ["../src/**/*.stories.mdx", "../src/**/*.stories.@(js|jsx|ts|tsx)"],
  addons: ["@storybook/addon-links", "@storybook/addon-essentials", "@storybook/addon-interactions", "@storybook/addon-mdx-gfm"],
  framework: {
    name: "@storybook/react-webpack5",
    options: {}
  },
  features: {
    previewMdx2: true
  },
  typescript: {
    check: false,
    checkOptions: {},
    reactDocgen: "react-docgen-typescript",
    reactDocgenTypescriptOptions: {
      // speeds up storybook build time
      allowSyntheticDefaultImports: false,
      // speeds up storybook build time
      esModuleInterop: false,
      // makes union prop types like variant and size appear as select controls
      shouldExtractLiteralValuesFromEnum: true,
      // makes string and boolean types that can be undefined appear as inputs and switches
      shouldRemoveUndefinedFromOptional: true,
      // Filter out third-party props from node_modules except @mui packages
      propFilter: prop => prop.parent ? !/node_modules\/(?!@mui)/.test(prop.parent.fileName) : true
    }
  },
  docs: {
    autodocs: true
  }
};