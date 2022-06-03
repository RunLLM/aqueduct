const path = require('path');

const withTM = require('next-transpile-modules')(["@aqueducthq/common"]);

module.exports = withTM({
    images: {
        domains: ['aqueduct-public-assets-bucket.s3.us-east-2.amazonaws.com'],
    },
    publicRuntimeConfig: {
        apiAddress: process.env.SERVER_ADDRESS,
        httpProtocol: process.env.NEXT_PUBLIC_PROTOCOL,
    },
    productionBrowserSourceMaps: true,
    webpack: (config, options) => {
        config.module.rules.push({
            test: /..\/common\/.*.ts$|..\/common\/.*.tsx$/,
            use: [options.defaultLoaders.babel],
        });

        // we want to ensure that the server project's version of react is used in all cases
        config.resolve.alias['react'] = path.join(__dirname, 'node_modules', 'react');
        config.resolve.alias['react-dom'] = path.resolve(__dirname, 'node_modules', 'react-dom');
        config.resolve.alias['react-redux'] = path.resolve(__dirname, 'node_modules', 'react-redux');

        // Make sure that we use the redux store defined here, not in ui-components.
        // Future work we'll just have redux state defined here.
        config.resolve.alias['@reduxjs/toolkit'] = path.resolve(__dirname, 'node_modules', '@reduxjs/toolkit');

        // Same logic here, we want to use our local version of emotion to avoid issues.
        config.resolve.alias['@emotion/react'] = path.resolve(__dirname, 'node_modules', '@emotion/react');
        config.resolve.alias['react-flow-renderer'] = path.resolve(__dirname, 'node_modules', 'react-flow-renderer');

        return config;
    },
});
