import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import { CSVConfig } from '../../../utils/integrations';
import React, { useEffect, useState } from 'react';

import { IntegrationFileUploadField } from './IntegrationFileUploadField';
import { IntegrationTextInputField } from './IntegrationTextInputField';

type Props = {
  setDialogConfig: (config: CSVConfig) => void;
};

export const CSVDialog: React.FC<Props> = ({ setDialogConfig }) => {
  const [name, setName] = useState<string>('');
  const [csv, setCSV] = useState(null);

  useEffect(() => {
    const config: CSVConfig = {
      name: name,
      csv: csv,
    };
    setDialogConfig(config);
  }, [name, csv]);

  return (
    <Box sx={{ mt: 2 }}>
      <Typography>Upload a CSV file to the demo database.</Typography>
      <IntegrationTextInputField
        label={'Table Name*'}
        description={'The name of the table to create.'}
        spellCheck={false}
        required={true}
        placeholder={name}
        onChange={(event) => setName(event.target.value)}
        value={name}
      />
      <IntegrationFileUploadField
        label={'CSV File*'}
        description={'The CSV file to populate the table in the demo database.'}
        required={true}
        placeholder={'Upload the CSV file.'}
        file={csv}
        onFile={(file) => {
          name ? null : setName(`${file.name.slice(0, -4)}`);
          const reader = new FileReader();
          reader.onloadend = function (event) {
            const content = event.target.result as string;
            setCSV({ name: file.name, data: content });
          };
          reader.readAsText(file);
        }}
        onReset={(_) => {
          setName('');
          setCSV(null);
        }}
      />
    </Box>
  );
};
