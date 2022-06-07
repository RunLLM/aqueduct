import '@aqueducthq/common/src/styles/globals.css';
import '@fortawesome/fontawesome-svg-core/styles.css';

import { theme } from '@aqueducthq/common/src/styles/theme/theme';
import { config } from '@fortawesome/fontawesome-svg-core';
import { createTheme, ThemeProvider } from '@mui/material/styles';
import { store } from '@stores/store';
import type { AppProps } from 'next/app';
import React from 'react';
import { Provider } from 'react-redux';

config.autoAddCss = false;

const Aqueduct: React.FC<AppProps> = ({ Component, pageProps }) => {
    const muiTheme = createTheme(theme);

    return (
        <Provider store={store}>
            <ThemeProvider theme={muiTheme}>
                <Component {...pageProps} />
            </ThemeProvider>
        </Provider>
    );
};

export default Aqueduct;
