import { Theme, ThemeOptions, Palette, PaletteOptions } from '@mui/material/styles';

declare module '@mui/material/styles' {
    interface PaletteShades {
        900?: string;
        800?: string;
        700?: string;
        600?: string;
        500?: string;
        400?: string;
        300?: string;
        200?: string;
        100?: string;
        75?: string;
        50?: string;
        25?: string;
    }

    interface Palette {
        black: string;
        white: string;
        darkGray: string;
        gray: PaletteShades;
        blue: PaletteShades;
        red: PaletteShades;
        green: PaletteShades;
        orange: PaletteShades;
        purple: PaletteShades;
        teal: PaletteShades;
        yellow: PaletteShades;
    }

    interface Theme {
        palette: Palette;
    }

    interface PaletteShadeOptions {
        900?: string;
        800?: string;
        700?: string;
        600?: string;
        500?: string;
        400?: string;
        300?: string;
        200?: string;
        100?: string;
        75?: string;
        50?: string;
        25?: string;
    }

    interface PaletteOptions {
        black?: string;
        white?: string;
        darkGray?: string;
        gray?: PaletteShadeOptions;
        blue?: PaletteShadeOptions;
        red?: PaletteShadeOptions;
        green?: PaletteShadeOptions;
        orange?: PaletteShadeOptions;
        purple?: PaletteShadeOptions;
        teal?: PaletteShadeOptions;
        yellow?: PaletteShadeOptions;
    }

    // This allows us to pass things into `createTheme` without a type error --
    // if we're adding more customization to our theme, we'll need to modify
    // both the `Theme` interface above and this.
    interface ThemeOptions {
        palette?: PaletteOptions;
    }
}




