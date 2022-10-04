import Box from '@mui/material/Box';
import React, { useState } from 'react';

import { Error, Logs } from '../../utils/shared';
import { Tab, Tabs } from '../primitives/Tabs.styles';

type Props = {
  logs?: Logs;
  err?: Error;
  contentHeight?: string;
};
const LogViewer: React.FC<Props> = ({ logs, err, contentHeight = '10vh' }) => {
  const hasOutput = (obj) => {
    return obj !== undefined && obj.length > 0;
  };

  const [selectedTab, setSelectedTab] = useState(0);
  const tabPanelOptions = {
    height: contentHeight,
    overflow: 'auto',
    fontFamily: 'monospace, monospace',
  };
  const errorTabPanelOptions = { ...tabPanelOptions, color: 'red.500' };
  const empty = { color: 'gray.200' };

  const noStdOut = 'No output logs to display.';
  const noStdErr = 'No error logs to display.';
  const noErr = 'No errors to display.';

  return (
    <Box sx={{ mb: 4 }} pb={1}>
      <Tabs
        value={selectedTab}
        onChange={(e, idx) => {
          e.preventDefault();
          setSelectedTab(idx);
        }}
        sx={{ mb: 1 }}
      >
        <Tab label="Errors" key="errors" />
        <Tab label="stdout" key="stdout" />
        <Tab label="stderr" key="strderr" />
      </Tabs>

      <Box
        key={0}
        role="tabpanel"
        sx={errorTabPanelOptions}
        hidden={selectedTab !== 0}
      >
        {err !== undefined && hasOutput(err?.tip)
          ? `${err.tip}:\n${err.context}`
          : noErr}
      </Box>
      <Box
        key={1}
        role="tabpanel"
        sx={tabPanelOptions}
        hidden={selectedTab !== 1}
      >
        {hasOutput(logs?.stdout) ? logs.stdout : noStdOut}
      </Box>
      <Box
        key={2}
        role="tabpanel"
        sx={errorTabPanelOptions}
        hidden={selectedTab !== 2}
      >
        {hasOutput(logs?.stderr) ? logs.stderr : noStdErr}
      </Box>
    </Box>
  );
};

export default LogViewer;
