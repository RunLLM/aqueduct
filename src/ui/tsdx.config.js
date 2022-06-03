const postcss = require('rollup-plugin-postcss');
const autoprefixer = require('autoprefixer');
const cssnano = require('cssnano');

module.exports = {
    rollup(config, options) {
        config.plugins.push(
            postcss({
                modules: true,
            })
        );
        return config;
    },
};

/*
module.exports = {
    rollup(config, options) {
        config.plugins.push(
            postcss({
                plugins: [
                    autoprefixer(),
                    cssnano({
                        preset: 'default',
                    }),
                ],
                inject: false,
                // only write out CSS for the first bundle (avoids pointless extra files):
                extract: !!options.writeMeta,
            })
        );
        return config;
    },
};*/