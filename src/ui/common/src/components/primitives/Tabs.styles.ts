import { styled } from '@mui/material/styles';
import Tab, { tabClasses } from '@mui/material/Tab';
import Tabs, { tabsClasses } from '@mui/material/Tabs';

const AqueductTabs = styled(Tabs)(({ theme }) => {
    return {
        [`& .${tabsClasses.indicator}`]: {
            backgroundColor: theme.palette.blue[800],
        },
    };
});

const AqueductTab = styled(Tab)(({ theme }) => {
    return {
        [`&.${tabClasses.selected}`]: {
            color: theme.palette.blue[800],
        },
    };
});

AqueductTab.defaultProps = {
    disableRipple: true,
};

export { AqueductTab as Tab, AqueductTabs as Tabs };
