import { createSlice, PayloadAction } from '@reduxjs/toolkit';

export interface OpenSideSheetState {
    leftSideSheetOpen: boolean;
    rightSideSheetOpen: boolean;
    bottomSideSheetOpen: boolean;
    workflowStatusBarOpen: boolean;
}

const initialOpenState: OpenSideSheetState = {
    leftSideSheetOpen: false,
    rightSideSheetOpen: false,
    bottomSideSheetOpen: false,
    workflowStatusBarOpen: true,
};

export const openSideSheetSlice = createSlice({
    name: 'openSideSheet',
    initialState: initialOpenState,
    reducers: {
        setLeftSideSheetOpenState: (state, { payload }: PayloadAction<boolean>) => {
            state.leftSideSheetOpen = payload;
        },
        setBottomSideSheetOpenState: (state, { payload }: PayloadAction<boolean>) => {
            state.bottomSideSheetOpen = payload;
        },
        setRightSideSheetOpenState: (state, { payload }: PayloadAction<boolean>) => {
            state.rightSideSheetOpen = payload;
        },
        setWorkflowStatusBarOpenState: (state, { payload }: PayloadAction<boolean>) => {
            state.workflowStatusBarOpen = payload;
        },
        setAllSideSheetState: (state, { payload }: PayloadAction<boolean>) => {
            state.leftSideSheetOpen = payload;
            state.bottomSideSheetOpen = payload;
            state.rightSideSheetOpen = payload;
        },
    },
});

export const {
    setLeftSideSheetOpenState,
    setRightSideSheetOpenState,
    setBottomSideSheetOpenState,
    setWorkflowStatusBarOpenState,
    setAllSideSheetState,
} = openSideSheetSlice.actions;

export default openSideSheetSlice.reducer;
