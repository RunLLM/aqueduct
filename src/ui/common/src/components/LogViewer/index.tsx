import Box from '@mui/material/Box';
import React, { useState } from 'react';

import { Error, Logs } from '../../utils/shared';
import { Tab, Tabs } from '../primitives/Tabs.styles';

type Props = {
  logs?: Logs;
  err?: Error;
};
const LogViewer: React.FC<Props> = ({ logs, err }) => {
  const hasOutput = (obj) => {
    return obj !== undefined && obj.length > 0;
  };

  const [selectedTab, setSelectedTab] = useState(0);
  const emptyElement = (
    <Box sx={{ p: 2, backgroundColor: 'gray.100' }}>Nothing to see here!</Box>
  );

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

      <Box key={0} role="tabpanel" hidden={selectedTab !== 0}>
        {err !== undefined && hasOutput(err?.tip) ? (
          <Box
            sx={{
              backgroundColor: 'red.100',
              color: 'red.600',
              p: 2,
              height: 'fit-content',
            }}
          >
            <pre style={{ margin: '0px' }}>{`${err.tip}\n${err.context}`}</pre>
          </Box>
        ) : (
          emptyElement
        )}
      </Box>

      <Box key={1} role="tabpanel" hidden={selectedTab !== 1}>
        {hasOutput(logs?.stdout) ? (
          <Box
            sx={{ backgroundColor: 'gray.100', p: 2, height: 'fit-content' }}
          >
            <pre style={{ margin: '0px' }}>{logs.stdout}</pre>
          </Box>
        ) : (
          emptyElement
        )}
      </Box>

      <Box key={2} role="tabpanel" hidden={selectedTab !== 2}>
        {hasOutput(logs?.stderr) ? (
          <Box
            sx={{ backgroundColor: 'gray.100', p: 2, height: 'fit-content' }}
          >
            <pre style={{ margin: '0px' }}>{logs.stderr}</pre>
          </Box>
        ) : (
          emptyElement
        )}
      </Box>
    </Box>
  );
};

export default LogViewer;
