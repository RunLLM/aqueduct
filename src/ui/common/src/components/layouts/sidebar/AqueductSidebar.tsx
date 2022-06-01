import {
  faChevronDown,
  faChevronLeft,
  faChevronRight,
  faChevronUp,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { Component, ReactElement } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import {
  setBottomSideSheetOpenState,
  setLeftSideSheetOpenState,
  setRightSideSheetOpenState,
} from '../../../reducers/openSideSheet';
import { RootState } from '../../../stores/store';
import { theme } from '../../../styles/theme/theme';
import { AllTransition, WidthTransition } from '../../../utils/shared';
import {
  CollapsedStatusBarWidthInPx,
  StatusBarWidthInPx,
} from '../../workflows/StatusBar';
import { MenuSidebarWidth } from '../menuSidebar';

export enum SidebarPosition {
  // opens from left to right
  left = 'LEFT',
  // opens from right to left
  right = 'RIGHT',
  // opens from bottom to top.
  bottom = 'BOTTOM',
}

export const VerticalSidebarWidthsFloats = [0.35, 0.28, 0.22];
export const VerticalSidebarWidths = ['35%', '28%', '22%'];
export const CollapsedSidebarWidthInPx = 50;
export const CollapsedSidebarHeightInPx = 50;
export const BottomSidebarMarginInPx = 25; // The amount of space on the left and the right of the bottom sidebar.
export const BottomSidebarHeightInPx = 400;
export const BottomSidebarHeaderHeightInPx = 50;

type Props = {
  zIndex?: number;
  position?: SidebarPosition;
  isSideSheetActive?: () => boolean;
  getSideSheetTitle: () => string;
  getSideSheetHeadingContent?: () => React.ReactElement;
  showWhenCollapsed?: boolean;
  children: ReactElement<any, any> | Component;
};

/**
 *
 * @param workflowStatusBarOpen Whether or not the workflow status bar is open.
 * @param baseWidth What should be treated as 100% width of the enclosing container.
 * @returns bottomSidesheetWidth The width of the bottom sidesheet, in either a single string, or an array for responsiveness.
 */
export const getBottomSideSheetWidth = (
  workflowStatusBarOpen: boolean,
  baseWidth = '100%'
): string | string[] => {
  return `calc(${baseWidth} - ${MenuSidebarWidth} - ${
    2 * BottomSidebarMarginInPx
  }px - ${getBottomSidesheetOffset(workflowStatusBarOpen)})`;
};

/**
 *
 * @param workflowStatusBarOpen Whether or not the workflow status bar is open.
 * @returns bottomSidesheetOffset The y offset from the bottom of the screen.
 */
export const getBottomSidesheetOffset = (
  workflowStatusBarOpen: boolean
): string => {
  if (workflowStatusBarOpen) {
    return `${StatusBarWidthInPx}px`;
  } else {
    return `${CollapsedStatusBarWidthInPx}px`;
  }
};

// NOTE: The layout management logic here assumes there can only be one left,
// right, and bottom side sheet open at any given time. The vertical side
// sheets are given precedence over the bottom sheets -- in other words, if
// there is a right side sheet and a bottom side sheet open at the same time,
// the right side sheet will be full height, and the width of the bottom side
// sheet will be reduced in accordance with that.
//
// NOTE: This current implementation isn't set up to work with multiple
// vertical side sheets. We probably want to think about some smarter mechanics
// if we get to that point in the future.
export const AqueductSidebar: React.FC<Props> = ({
  zIndex = 1,
  children,
  position,
  getSideSheetTitle,
  getSideSheetHeadingContent = () => <div />,
  showWhenCollapsed = true,
  isSideSheetActive = () => true,
}) => {
  if (!isSideSheetActive()) {
    return null;
  }

  const dispatch = useDispatch();
  const openSideSheetState = useSelector(
    (state: RootState) => state.openSideSheetReducer
  );

  const bottomSideSheetWidth = getBottomSideSheetWidth(
    openSideSheetState.workflowStatusBarOpen
  );
  const bottomSideSheetOffset = getBottomSidesheetOffset(
    openSideSheetState.workflowStatusBarOpen
  );

  // TODO: (agiron123): Make these into styled components.
  const styles = {
    // open right to left
    rightAligned: {
      position: 'fixed',
      right: '0px',
      top: '0px',
      transition: WidthTransition,
      width: VerticalSidebarWidths,
      height: '100%',
      borderTop: '0px',
      borderLeft: '1px',
      borderRight: '0px',
      borderBottom: '0px',
      borderColor: theme.palette.gray['500'],
      borderStyle: 'solid',
    },
    // open left to right
    leftAligned: {
      position: 'fixed',
      left: '200px',
      top: '0px',
      transition: WidthTransition,
      width: VerticalSidebarWidths,
      height: '100%',
      borderTop: '0px',
      borderLeft: '0px',
      borderRight: '1px',
      borderBottom: '0px',
      borderColor: theme.palette.gray['500'],
      borderStyle: 'solid',
    },
    // open bottom to top
    bottomAligned: {
      position: 'absolute',
      right: bottomSideSheetOffset,
      bottom: 0,
      mx: `${BottomSidebarMarginInPx}px`,
      transition: AllTransition,
      width: bottomSideSheetWidth,
      height: `${BottomSidebarHeightInPx}px`,
      borderTop: '1px',
      borderLeft: '1px',
      borderRight: '1px',
      borderBottom: '0px',
      borderColor: theme.palette.gray['500'],
      borderStyle: 'solid',
    },
    propertiesText: { writingMode: 'vertical-rl', alignSelf: 'center' },
    sidebarIcon: { cursor: 'pointer' },
  };

  let sidebarAlignment, isOpen, onClickFn, ExpandIcon, CollapseIcon;

  switch (position) {
    case SidebarPosition.left: {
      sidebarAlignment = styles.leftAligned;
      ExpandIcon = faChevronRight;
      CollapseIcon = faChevronLeft;
      isOpen = openSideSheetState.leftSideSheetOpen;
      onClickFn = () => {
        dispatch(setLeftSideSheetOpenState(!isOpen));
      };
      break;
    }
    case SidebarPosition.right: {
      sidebarAlignment = styles.rightAligned;
      ExpandIcon = faChevronLeft;
      CollapseIcon = faChevronRight;
      isOpen = openSideSheetState.rightSideSheetOpen;
      onClickFn = () => {
        dispatch(setRightSideSheetOpenState(!isOpen));
      };
      break;
    }
    case SidebarPosition.bottom: {
      sidebarAlignment = styles.bottomAligned;
      ExpandIcon = faChevronUp;
      CollapseIcon = faChevronDown;
      isOpen = openSideSheetState.bottomSideSheetOpen;
      onClickFn = () => {
        dispatch(setBottomSideSheetOpenState(!isOpen));
      };
      break;
    }
    default: {
      sidebarAlignment = styles.rightAligned;
      break;
    }
  }

  let collapsedHeadingStyle, collapsedBoxStyle;
  if (position === SidebarPosition.bottom) {
    collapsedHeadingStyle = {};
    collapsedBoxStyle = {
      ...sidebarAlignment,
      display: 'flex',
      alignItems: 'center',
      height: `${CollapsedSidebarHeightInPx}px`,
      zIndex: zIndex,
      paddingLeft: '12px',
    };
  } else {
    collapsedHeadingStyle = {
      ...styles.propertiesText,
    };
    collapsedBoxStyle = {
      ...sidebarAlignment,
      width: `${CollapsedSidebarWidthInPx}px`,
      height: '100%',
      zIndex: zIndex,
      paddingTop: '12px',
    };
  }

  let CollapsedUI = null;
  if (position !== SidebarPosition.bottom && showWhenCollapsed) {
    CollapsedUI = (
      <Box
        sx={{
          backgroundColor: theme.palette.gray['100'],
          ...collapsedBoxStyle,
          cursor: 'pointer',
          display: 'flex',
          alignItems: 'center',
          flexDirection: 'column',
        }}
        onClick={onClickFn}
      >
        <Box sx={styles.sidebarIcon}>
          <FontAwesomeIcon icon={ExpandIcon} />
        </Box>
        <Typography variant="h5" sx={collapsedHeadingStyle} mt={1}>
          {getSideSheetTitle()}
        </Typography>
      </Box>
    );
  } else if (showWhenCollapsed) {
    CollapsedUI = (
      <Box
        sx={{
          backgroundColor: theme.palette.gray['100'],
          ...collapsedBoxStyle,
          display: 'flex',
          cursor: 'pointer',
          alignItems: 'center',
          height: BottomSidebarHeaderHeightInPx,
        }}
        onClick={onClickFn}
      >
        <Box sx={{ ...styles.sidebarIcon, m: 2 }}>
          <FontAwesomeIcon icon={ExpandIcon} />
        </Box>
        <Typography variant="h5" sx={collapsedHeadingStyle} ml={1}>
          {getSideSheetTitle()}
        </Typography>
      </Box>
    );
  }

  const ExpandedUI = (
    <Box
      sx={{
        ...sidebarAlignment,
        zIndex: zIndex,
        overflow: 'auto',
      }}
    >
      <>
        <Box
          sx={{
            display: 'flex',
            alignItems: 'center',
            height:
              position === SidebarPosition.bottom
                ? BottomSidebarHeaderHeightInPx
                : '',
            backgroundColor: theme.palette.gray['100'],
          }}
          py={1}
        >
          <Box sx={{ display: 'flex', flex: 1, alignItems: 'center' }}>
            <Box sx={{ cursor: 'pointer', m: 1 }} onClick={onClickFn}>
              <FontAwesomeIcon icon={CollapseIcon} />
            </Box>
            <Typography variant="h5">{getSideSheetTitle()}</Typography>
          </Box>

          <Box sx={{ mx: 2 }}>{getSideSheetHeadingContent()}</Box>
        </Box>

        {children}
      </>
    </Box>
  );

  return !isOpen ? CollapsedUI : ExpandedUI;
};

export default AqueductSidebar;
