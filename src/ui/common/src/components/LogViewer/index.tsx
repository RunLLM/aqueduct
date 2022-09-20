import { TabContext, TabList, TabPanel } from '@mui/lab';
import { Tabs, Tab } from '@mui/material';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import React, { useState } from 'react';

import { Error, Logs } from '../../utils/shared';
import LogBlock, { LogLevel } from '../text/LogBlock';

type Props = {
  logs?: Logs;
  err?: Error;
  contentHeight?: string;
};
const LogViewer: React.FC<Props> = ({ logs, err, contentHeight="10vh" }) => {
  const [currentTab, setCurrentTab] = useState("1");
  const tabPanelOptions = {height:contentHeight, overflow:"scroll", fontFamily:'monospace, monospace'};
  const errorTabPanelOptions = {...tabPanelOptions, color:'red.500'};
  if (err) {
    setCurrentTab("3");
  }
  return (
    <Box sx={{mb: 4}} pb={1}>
      <TabContext value={currentTab}>
        <Box sx={{ borderBottom: 1, borderColor: 'divider'}}>
          <TabList onChange={(_, tab) => setCurrentTab(tab)} aria-label="log viewer">
            <Tab label="Stdout" value="1" />
            <Tab label="Stderr" value="2" />
            <Tab label="Errors" value="3" />
          </TabList>
        </Box>
        <TabPanel sx={tabPanelOptions} value="1">{logs.stdout? logs.stdout: "No output to display."}</TabPanel>
        <TabPanel sx={errorTabPanelOptions} value="2">{logs.stderr? logs.stderr: "No error logs to display."}</TabPanel>
        <TabPanel sx={errorTabPanelOptions} value="3">{err? `${err.tip}:\n${err.context}`: "No errors."}</TabPanel>
      </TabContext>
    </Box>
  );
};

export default LogViewer;
