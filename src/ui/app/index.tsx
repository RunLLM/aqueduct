import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { HomePage, DataPage, IntegrationsPage, IntegrationDetailsPage, WorkflowPage, WorkflowsPage, LoginPage, ErrorPage, AccountPage, OperatorDetailsPage, ArtifactDetailsPage, MetricDetailsPage, CheckDetailsPage } from '@aqueducthq/common';
import { store } from './stores/store';
import { Provider } from 'react-redux';
import { useUser, UserProfile } from '@aqueducthq/common';
import { getPathPrefix } from '@aqueducthq/common/src/utils/getPathPrefix';
import '@aqueducthq/common/src/styles/globals.css';
import { createTheme, Palette, ThemeProvider } from '@mui/material/styles';

interface IPalette extends Palette {
  black: string;
  white: string;
  darkGray: string;
  gray: {
    900: string,
    800: string,
    700: string,
    600: string,
    500: string,
    400: string,
    300: string,
    200: string,
    100: string,
    90: string,
    75: string,
    50: string,
    25: string,
  };
  blue: {
    900: string,
    800: string,
    700: string,
    500: string,
    400: string,
    300: string,
    200: string,
    100: string,
    50: string,
  };
  red: {
    800: string,
    700: string,
    600: string,
    500: string,
    300: string,
    100: string,
    25: string,
  };
  green: {
    900: string,
    800: string,
    700: string,
    600: string,
    500: string,
    400: string,
    300: string,
    200: string,
    100: string,
    25: string,
  };
  orange: {
    700: string,
    600: string,
    500: string,
    100: string,
    25: string,
  };
  purple: {
    600: string,
    100: string,
  };
  teal: {
    800: string,
    100: string,
  };
  yellow: {
    800: string,
    500: string,
    100: string,
  };
  Info: string;
  Success: string;
  Warning: string;
  Error: string;
  Secondary: string;
  Primary: string;
  Default: string;
  Running: string;
  TableSuccessBackground: string;
  TableErrorBackground: string;
  TableWarningBackground: string;
  DarkContrast: string;
  DarkContrast50: string;
  DarkErrorMain: string;
  DarkErrorMain75: string;
  DarkErrorMain50: string;
  DarkWarningMain: string;
  DarkWarningMain75: string;
  DarkWarningMain50: string;
  DarkSuccessMain: string;
  DarkSuccessMain75: string;
  DarkSuccessMain50: string;
  LogoDarkBlue: string;
  LogoLight: string; //Matches logo color
  NavBackgroundDark: string; //(0,47,93)
  NavMenuHover: string; //(109,148,253)
  NavMenuActive: string; //(73,122,250)
};

// interface ITheme extends Theme {
//   palette: IPalette;
// }

// interface IThemeOptions extends ThemeOptions {
//   palette: IPalette;
// }


// declare module '@mui/material/styles' {
//   interface Theme {
//     red: {
//       800: React.CSSProperties['color'],
//       700: React.CSSProperties['color'],
//       600: React.CSSProperties['color'],
//       500: React.CSSProperties['color'],
//       300: React.CSSProperties['color'],
//       100: React.CSSProperties['color'],
//       25: React.CSSProperties['color'],
//     },
//   }

//   interface PaletteColor {
//     red?: PaletteOptions
//   }

//   interface ThemeOptions {
//     red?: PaletteOptions
//   }

//   interface PaletteOptions {
//     red: PaletteOptions['primary'];
//   }
// }

declare module '@mui/material/styles' {
  interface Theme {
    status: {
      danger: React.CSSProperties['color'];
    };
  }

  interface Palette {
    neutral: Palette['primary'];
  }

  interface PaletteOptions {
    neutral: PaletteOptions['primary'];
  }

  interface PaletteColor {
    darker?: string;
  }

  interface SimplePaletteColorOptions {
    darker?: string;
  }

  interface ThemeOptions {
    status: {
      danger: React.CSSProperties['color'];
    };
  }
}


// declare module "@mui/material/styles/createPalette" {
//   // This defines all the available keys that we have for our color palette.
//   export interface PaletteOptions {
//     // chip: {
//     //   color: string;
//     //   expandIcon: {
//     //     background: string;
//     //     color: string;
//     //   };
//     // };
//     black: string;
//     white: string;
//     darkGray: string;
//     gray: {
//       900: string,
//       800: string,
//       700: string,
//       600: string,
//       500: string,
//       400: string,
//       300: string,
//       200: string,
//       100: string,
//       90: string,
//       75: string,
//       50: string,
//       25: string,
//     };
//     blue: {
//       900: string,
//       800: string,
//       700: string,
//       500: string,
//       400: string,
//       300: string,
//       200: string,
//       100: string,
//       50: string,
//     };
//     red: {
//       800: string,
//       700: string,
//       600: string,
//       500: string,
//       300: string,
//       100: string,
//       25: string,
//     };
//     green: {
//       900: string,
//       800: string,
//       700: string,
//       600: string,
//       500: string,
//       400: string,
//       300: string,
//       200: string,
//       100: string,
//       25: string,
//     };
//     orange: {
//       700: string,
//       600: string,
//       500: string,
//       100: string,
//       25: string,
//     };
//     purple: {
//       600: string,
//       100: string,
//     };
//     teal: {
//       800: string,
//       100: string,
//     };
//     yellow: {
//       800: string,
//       500: string,
//       100: string,
//     };
//     Info: string;
//     Success: string;
//     Warning: string;
//     Error: string;
//     Secondary: string;
//     Primary: string;
//     Default: string;
//     Running: string;
//     TableSuccessBackground: string;
//     TableErrorBackground: string;
//     TableWarningBackground: string;
//     DarkContrast: string;
//     DarkContrast50: string;
//     DarkErrorMain: string;
//     DarkErrorMain75: string;
//     DarkErrorMain50: string;
//     DarkWarningMain: string;
//     DarkWarningMain75: string;
//     DarkWarningMain50: string;
//     DarkSuccessMain: string;
//     DarkSuccessMain75: string;
//     DarkSuccessMain50: string;
//     LogoDarkBlue: string;
//     LogoLight: string; //Matches logo color
//     NavBackgroundDark: string; //(0,47,93)
//     NavMenuHover: string; //(109,148,253)
//     NavMenuActive: string; //(73,122,250)
//   }
// }

// export interface PaletteOptions {
//   primary?: PaletteColorOptions;
//   secondary?: PaletteColorOptions;
//   error?: PaletteColorOptions;
//   warning?: PaletteColorOptions;
//   info?: PaletteColorOptions;
//   success?: PaletteColorOptions;
//   mode?: PaletteMode;
//   tonalOffset?: PaletteTonalOffset;
//   contrastThreshold?: number;
//   common?: Partial<CommonColors>;
//   grey?: ColorPartial;
//   text?: Partial<TypeText>;
//   divider?: string;
//   action?: Partial<TypeAction>;
//   background?: Partial<TypeBackground>;
//   getContrastText?: (background: string) => string;
// }

// TODO: Figure out if this augmentation is needed. My guess is not.
// declare module '@mui/material/styles' {
//   interface Theme {
//     palette: {
//       black: string;
//       white: string;
//       darkGray: string;
//       gray: {
//         900: string,
//         800: string,
//         700: string,
//         600: string,
//         500: string,
//         400: string,
//         300: string,
//         200: string,
//         100: string,
//         90: string,
//         75: string,
//         50: string,
//         25: string,
//       },
//       blue: {
//         900: string,
//         800: string,
//         700: string,
//         500: string,
//         400: string,
//         300: string,
//         200: string,
//         100: string,
//         50: string,
//       },
//       red: {
//         800: string,
//         700: string,
//         600: string,
//         500: string,
//         300: string,
//         100: string,
//         25: string,
//       },
//       green: {
//         900: string,
//         800: string,
//         700: string,
//         600: string,
//         500: string,
//         400: string,
//         300: string,
//         200: string,
//         100: string,
//         25: string,
//       },
//       orange: {
//         700: string,
//         600: string,
//         500: string,
//         100: string,
//         25: string,
//       },
//       purple: {
//         600: string,
//         100: string,
//       },
//       teal: {
//         800: string,
//         100: string,
//       },
//       yellow: {
//         800: string,
//         500: string,
//         100: string,
//       },
//       Info: string,
//       Success: string,
//       Warning: string,
//       Error: string,
//       Secondary: string,
//       Primary: string,
//       Default: string,
//       Running: string,
//       TableSuccessBackground: string,
//       TableErrorBackground: string,
//       TableWarningBackground: string,
//       DarkContrast: string,
//       DarkContrast50: string,
//       DarkErrorMain: string,
//       DarkErrorMain75: string,
//       DarkErrorMain50: string,
//       DarkWarningMain: string,
//       DarkWarningMain75: string,
//       DarkWarningMain50: string,
//       DarkSuccessMain: string,
//       DarkSuccessMain75: string,
//       DarkSuccessMain50: string,
//       LogoDarkBlue: string,
//       LogoLight: string, //Matches logo color
//       NavBackgroundDark: string, //(0,47,93)
//       NavMenuHover: string, //(109,148,253)
//       NavMenuActive: string, //(73,122,250)
//     }
//   }
// }



function RequireAuth({ children, user }): { children: JSX.Element, user: UserProfile | undefined } {
  const pathPrefix = getPathPrefix();

  if (!user || !user.apiKey) {
    return <Navigate to={`${pathPrefix}/login`} replace />;
  }

  return children;
}

const App = () => {
  const { user, loading } = useUser();
  if (loading) {
    return null;
  }

  const pathPrefix = getPathPrefix();
  let routesContent: React.ReactElement;
  routesContent = (
    <Routes>
      <Route path={`${pathPrefix ?? "/"}`} element={<RequireAuth user={user}><HomePage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/data`} element={<RequireAuth user={user}><DataPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/integrations`} element={<RequireAuth user={user}><IntegrationsPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/integration/:id`} element={<RequireAuth user={user}><IntegrationDetailsPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/workflows`} element={<RequireAuth user={user}><WorkflowsPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/login`} element={user && user.apiKey ? <Navigate to="/" replace /> : <LoginPage />} />
      <Route path={`/${pathPrefix}/account`} element={<RequireAuth user={user}><AccountPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/workflow/:id`} element={<RequireAuth user={user}><WorkflowPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/workflow/:workflowId/result/:workflowDagResultId/operator/:operatorId`} element={<RequireAuth user={user}><OperatorDetailsPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/workflow/:workflowId/result/:workflowDagResultId/artifact/:artifactId`} element={<RequireAuth user={user}><ArtifactDetailsPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/workflow/:workflowId/result/:workflowDagResultId/metric/:metricOperatorId`} element={<RequireAuth user={user}><MetricDetailsPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/workflow/:workflowId/result/:workflowDagResultId/check/:checkOperatorId`} element={<RequireAuth user={user}><CheckDetailsPage user={user} /> </RequireAuth>} />
      <Route path={`/${pathPrefix}/404`} element={user && user.apiKey ? <RequireAuth user={user}><ErrorPage user={user} /> </RequireAuth> : <ErrorPage />} />
      <Route path="*" element={<Navigate replace to={`/404`} />} />
    </Routes>
  );

  // const muiTheme = createTheme(
  //   {
  //     palette: {
  //       black: '#000000',
  //       white: '#ffffff',
  //       darkGray: '#333333',
  //       gray: {
  //         900: '#101840',
  //         800: '#474d66',
  //         700: '#696f8c',
  //         600: '#8f95b2',
  //         500: '#c1c4d6',
  //         400: '#d8dae5',
  //         300: '#E6E8F0',
  //         200: '#edeff5',
  //         100: '#F4F5F9',
  //         90: '#F4F6FA',
  //         75: '#F9FAFC',
  //         50: '#F2F2F2',
  //         25: '#F9F9F9',
  //       },
  //       blue: {
  //         900: '#002F5E',
  //         800: '#004080',
  //         700: '#0059B3',
  //         500: '#0073E6',
  //         400: '#0080FF',
  //         300: '#4DA6FF',
  //         200: '#66B3FF',
  //         100: '#CCE6FF',
  //         50: '#E6F2FF',
  //       },
  //       red: {
  //         800: '#611F1F',
  //         700: '#7D2828',
  //         600: '#A73636',
  //         500: '#D14343',
  //         300: '#EE9191',
  //         100: '#F9DADA',
  //         25: '#FDF4F4',
  //       },
  //       green: {
  //         900: '#10261E',
  //         800: '#214C3C',
  //         700: '#317159',
  //         600: '#429777',
  //         500: '#52BD95',
  //         400: '#75CAAA',
  //         300: '#97D7BF',
  //         200: '#BAE5D5',
  //         100: '#DCF2EA',
  //         25: '#F5FBF8',
  //       },
  //       orange: {
  //         700: '#996A13',
  //         600: '#FFA600',
  //         500: '#FFB020',
  //         100: '#F8E3DA',
  //         25: '#FFFAF2',
  //       },
  //       purple: {
  //         600: '#6E62B6',
  //         100: '#E7E4F9',
  //       },
  //       teal: {
  //         800: '#0F5156',
  //         100: '#D3F5F7',
  //       },
  //       yellow: {
  //         800: '#66460D',
  //         500: '#FFB833',
  //         100: '#FFEFD2',
  //       },
  //       Info: '#0288D1',
  //       Success: '#2e7d32',
  //       Warning: '#ed6c02',
  //       Error: '#d32f2f',
  //       Secondary: '#9c27b0',
  //       Primary: '#1976d2',
  //       Default: '#8f95b2',
  //       Running: '#00FFFF',
  //       TableSuccessBackground: 'rgba(76,175,80,0.1)',
  //       TableErrorBackground: 'rgba(224,67,54,0.1)',
  //       TableWarningBackground: 'rgba(237,18,2,0.1)',
  //       DarkContrast: 'rgba(0, 0, 0, 1)',
  //       DarkContrast50: 'rgba(0, 0, 0, 0.50)',
  //       DarkErrorMain: 'rgba(244, 67, 54, 1)',
  //       DarkErrorMain75: 'rgba(244, 67, 54, 0.75)',
  //       DarkErrorMain50: 'rgba(244, 67, 54, 0.5)',
  //       DarkWarningMain: 'rgba(255, 167, 30, 1)',
  //       DarkWarningMain75: 'rgba(255, 167, 30, 0.75)',
  //       DarkWarningMain50: 'rgba(255, 167, 30, 0.5)',
  //       DarkSuccessMain: 'rgba(102, 187, 106, 1)',
  //       DarkSuccessMain75: 'rgba(102, 187, 106, 0.75)',
  //       DarkSuccessMain50: 'rgba(102, 187, 106, 0.5)',
  //       LogoDarkBlue: '#002F5E',
  //       LogoLight: '#9CE1FF', //Matches logo color
  //       NavBackgroundDark: '#002F5D', //(0,47,93)
  //       NavMenuHover: '#6D94FD', //(109,148,253)
  //       NavMenuActive: '#497AFA', //(73,122,250)
  //     }
  //   }
  // );



  // TODO: Figure out how to do this without having to define all values here.
  // const themeOptions: IThemeOptions = {
  //   palette: {
  //     // values from default theme:
  //     // https://mui.com/material-ui/customization/default-theme/
  //     common: {
  //       black: '#000',
  //       white: '#fff',
  //     },
  //     mode: "light",
  //     contrastThreshold: 3,
  //     tonalOffset: 0.2,
  //     primary: {
  //       main: '#1976d2',
  //       light: '#42a5f5',
  //       dark: '#1565c0',
  //       contrastText: '#fff'
  //     },
  //     secondary: {
  //       main: '#9c27b0',
  //       light: '#ba68c8',
  //       dark: '#7b1fa2',
  //       contrastText: '#fff'
  //     },
  //     error: {
  //       main: '#d32f2f',
  //       light: '#ef5350',
  //       dark: '#c62828',
  //       contrastText: '#fff'
  //     },
  //     warning: {
  //       main: '#ed6c02',
  //       light: '#ff9800',
  //       dark: '#e65100',
  //       contrastText: '#fff'
  //     },
  //     info: {
  //       main: '#0288d1',
  //       light: '#03a9f4',
  //       dark: '#01579b',
  //       contrastText: '#fff'
  //     },
  //     success: {
  //       main: '#2e7d32',
  //       light: '#4caf50',
  //       dark: '#1b5e20',
  //       contrastText: '#fff'
  //     },
  //     grey: {
  //       50: '#fafafa',
  //       100: '#f5f5f5',
  //       200: '#eeeeee',
  //       300: '#e0e0e0',
  //       400: '#bdbdbd',
  //       500: '#9e9e9e',
  //       600: '#757575',
  //       700: '#616161',
  //       800: '#424242',
  //       900: '#212121',
  //       'A100': '#f5f5f5',
  //       'A200': '#eeeeee',
  //       'A400': '#bdbdbd',
  //       'A700': '#616161'
  //     },
  //     text: {
  //       primary: 'rgba(0,0,0,0.87)',
  //       secondary: 'rgba(0,0,0,0.6)',
  //       disabled: 'rgba(0,0,0,0.38)',
  //     },
  //     divider: 'rgba(0,0,0,0.12)',
  //     background: {
  //       default: '#fff',
  //       paper: '#fff'
  //     },
  //     action: {
  //       active: 'rgba(0,0,0,0.54)',
  //       hover: 'rgba(0,0,0,0.04)',
  //       hoverOpacity: 0.04,
  //       selected: 'rgba(0,0,0,0.08)',
  //       selectedOpacity: 0.08,
  //       disabled: 'rgba(0,0,0,0.26)',
  //       disabledBackground: 'rgba(0,0,0,0.12)',
  //       disabledOpacity: 0.38,
  //       focus: 'rgba(0,0,0,0.12)',
  //       focusOpacity: 0.12,
  //       activatedOpacity: 0.12
  //     },
  //     // TODO: Figure out what this is suppoed to be. Not sure when we're using this.
  //     // getContrastText: () => {
  //     //   console.log('getContrastText');
  //     //   return '#fff'
  //     // },
  //     // augmentColor: (option) => {
  //     //   return '#fff'
  //     // },
  //     // end values from default theme
  //     // values from aqueduct theme:
  //     black: '#000000',
  //     white: '#ffffff',
  //     darkGray: '#333333',
  //     gray: {
  //       900: '#101840',
  //       800: '#474d66',
  //       700: '#696f8c',
  //       600: '#8f95b2',
  //       500: '#c1c4d6',
  //       400: '#d8dae5',
  //       300: '#E6E8F0',
  //       200: '#edeff5',
  //       100: '#F4F5F9',
  //       90: '#F4F6FA',
  //       75: '#F9FAFC',
  //       50: '#F2F2F2',
  //       25: '#F9F9F9',
  //     },
  //     blue: {
  //       900: '#002F5E',
  //       800: '#004080',
  //       700: '#0059B3',
  //       500: '#0073E6',
  //       400: '#0080FF',
  //       300: '#4DA6FF',
  //       200: '#66B3FF',
  //       100: '#CCE6FF',
  //       50: '#E6F2FF',
  //     },
  //     red: {
  //       800: '#611F1F',
  //       700: '#7D2828',
  //       600: '#A73636',
  //       500: '#D14343',
  //       300: '#EE9191',
  //       100: '#F9DADA',
  //       25: '#FDF4F4',
  //     },
  //     green: {
  //       900: '#10261E',
  //       800: '#214C3C',
  //       700: '#317159',
  //       600: '#429777',
  //       500: '#52BD95',
  //       400: '#75CAAA',
  //       300: '#97D7BF',
  //       200: '#BAE5D5',
  //       100: '#DCF2EA',
  //       25: '#F5FBF8',
  //     },
  //     orange: {
  //       700: '#996A13',
  //       600: '#FFA600',
  //       500: '#FFB020',
  //       100: '#F8E3DA',
  //       25: '#FFFAF2',
  //     },
  //     purple: {
  //       600: '#6E62B6',
  //       100: '#E7E4F9',
  //     },
  //     teal: {
  //       800: '#0F5156',
  //       100: '#D3F5F7',
  //     },
  //     yellow: {
  //       800: '#66460D',
  //       500: '#FFB833',
  //       100: '#FFEFD2',
  //     },
  //     Info: '#0288D1',
  //     Success: '#2e7d32',
  //     Warning: '#ed6c02',
  //     Error: '#d32f2f',
  //     Secondary: '#9c27b0',
  //     Primary: '#1976d2',
  //     Default: '#8f95b2',
  //     Running: '#00FFFF',
  //     TableSuccessBackground: 'rgba(76,175,80,0.1)',
  //     TableErrorBackground: 'rgba(224,67,54,0.1)',
  //     TableWarningBackground: 'rgba(237,18,2,0.1)',
  //     DarkContrast: 'rgba(0, 0, 0, 1)',
  //     DarkContrast50: 'rgba(0, 0, 0, 0.50)',
  //     DarkErrorMain: 'rgba(244, 67, 54, 1)',
  //     DarkErrorMain75: 'rgba(244, 67, 54, 0.75)',
  //     DarkErrorMain50: 'rgba(244, 67, 54, 0.5)',
  //     DarkWarningMain: 'rgba(255, 167, 30, 1)',
  //     DarkWarningMain75: 'rgba(255, 167, 30, 0.75)',
  //     DarkWarningMain50: 'rgba(255, 167, 30, 0.5)',
  //     DarkSuccessMain: 'rgba(102, 187, 106, 1)',
  //     DarkSuccessMain75: 'rgba(102, 187, 106, 0.75)',
  //     DarkSuccessMain50: 'rgba(102, 187, 106, 0.5)',
  //     LogoDarkBlue: '#002F5E',
  //     LogoLight: '#9CE1FF', //Matches logo color
  //     NavBackgroundDark: '#002F5D', //(0,47,93)
  //     NavMenuHover: '#6D94FD', //(109,148,253)
  //     NavMenuActive: '#497AFA', //(73,122,250)
  //   }
  // }

  //const muiTheme = createTheme(themeOptions);

  //console.log('muiTheme: ', muiTheme);
  // const latestTheme = createTheme({
  //   palette: {
  //     red: {
  //       800: '#611F1F',
  //       700: '#7D2828',
  //       600: '#A73636',
  //       500: '#D14343',
  //       300: '#EE9191',
  //       100: '#F9DADA',
  //       // 25: '#FDF4F4',
  //     },
  //   }
  // });


  const theme = createTheme({
    status: {
      danger: '#e53e3e',
    },
    palette: {
      primary: {
        main: '#0971f1',
        darker: '#053e85',
      },
      neutral: {
        main: '#64748B',
        contrastText: '#fff',
      },
    },
  });

  console.log('themeBefore passing to context: ', theme);

  return (
    <ThemeProvider theme={theme}>
      <BrowserRouter>{routesContent}</BrowserRouter>
    </ThemeProvider>
  );
};

const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);

const theme = createTheme({
  status: {
    danger: '#e53e3e',
  },
  palette: {
    primary: {
      main: '#0971f1',
      darker: '#053e85',
    },
    neutral: {
      main: '#64748B',
      contrastText: '#fff',
    },
  },
});

root.render(
  <ThemeProvider theme={theme}>
    <Provider store={store}>
      <App />
    </Provider>
  </ThemeProvider>
);
