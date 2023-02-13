import React from 'react';

interface TabPanelProps {
  children?: React.ReactNode;
  index: string;
  value: string;
}

// TODO: Make this a component, probably can put this near the other tab component that we have
// Linear Task: https://linear.app/aqueducthq/issue/ENG-2409/tabpanel-component
export const TabPanel: React.FC<TabPanelProps> = (props: TabPanelProps) => {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`simple-tabpanel-${index}`}
      aria-labelledby={`simple-tab-${index}`}
      {...other}
      style={{
        height: '100%',
        width: '100%',
        margin: 0,
        padding: '0 0 32px 0',
      }}
    >
      {value === index && children}
    </div>
  );
};

export default TabPanel;
