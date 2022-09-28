import { TabContext, TabList, TabPanel } from '@mui/lab';
import { Tab } from '@mui/material';
import Box from '@mui/material/Box';
import React, { useState } from 'react';

import { Error, Logs } from '../../utils/shared';

type Props = {
  logs?: Logs;
  err?: Error;
  contentHeight?: string;
};
const LogViewer: React.FC<Props> = ({ logs, err, contentHeight = '10vh' }) => {
  const hasOutput = (obj) => {
    return obj !== undefined && obj.length > 0;
  };

  const [currentTab, setCurrentTab] = useState('1');

  console.log(currentTab);

  const tabPanelOptions = {
    height: contentHeight,
    overflow: 'scroll',
    fontFamily: 'monospace, monospace',
  };
  const errorTabPanelOptions = { ...tabPanelOptions, color: 'red.500' };
  const empty = { color: 'gray.200' };

  const noStdOut = 'No output logs to display.';
  const noStdErr = 'No error logs to display.';
  const noErr = 'No errors to display.';

  return (
    <Box sx={{ mb: 4 }} pb={1}>
      <TabContext value={currentTab}>
        <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
          <TabList
            onChange={(_, tab) => setCurrentTab(tab)}
            aria-label="log viewer"
          >
            <Tab
              sx={hasOutput(logs?.stdout) ? {} : empty}
              label="Stdout"
              value="1"
            />
            <Tab
              sx={hasOutput(logs?.stderr) ? {} : empty}
              label="Stderr"
              value="2"
            />
            <Tab
              sx={err !== undefined && hasOutput(err?.tip) ? {} : empty}
              label="Errors"
              value="3"
            />
          </TabList>
        </Box>

        <TabPanel sx={tabPanelOptions} value="1">
          {hasOutput(logs?.stdout) ? logs.stdout : noStdOut}
        </TabPanel>
        <TabPanel sx={errorTabPanelOptions} value="2">
          {hasOutput(logs?.stderr) ? logs.stderr : noStdErr}
        </TabPanel>
        <TabPanel sx={errorTabPanelOptions} value="3">
          {err !== undefined && hasOutput(err?.tip)
            ? `${err.tip}:\n${err.context}`
            : noErr}
        </TabPanel>
      </TabContext>
    </Box>
  );
};

export default LogViewer;
