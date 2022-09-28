import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import { DataGrid } from '@mui/x-data-grid';
import React, { useEffect, useState } from 'react';

import { CSVConfig, FileData } from '../../../utils/integrations';
import { IntegrationFileUploadField } from './IntegrationFileUploadField';
import { IntegrationTextInputField } from './IntegrationTextInputField';

type Props = {
  setDialogConfig: (config: CSVConfig) => void;
  setErrMsg: (msg: string) => void;
};

export const CSVDialog: React.FC<Props> = ({ setDialogConfig, setErrMsg }) => {
  const [name, setName] = useState<string>('');
  const [csv, setCSV] = useState(null);

  useEffect(() => {
    const config: CSVConfig = {
      name: name,
      csv: csv,
    };
    setDialogConfig(config);
  }, [name, csv]);

  const displayFileFn = (file: FileData) => {
    const allRows = file.data.split(/\r?\n/);
    const parsedHeader = ['id'];
    parsedHeader.push(...allRows[0].split(/,/));

    const width = 25;
    const parsedColumns = parsedHeader.map((headerName) => {
      let hideColumn = false;
      if (headerName === 'id') {
        hideColumn = true;
      }

      return {
        field: headerName,
        headerName: headerName,
        width: width * headerName.length,
        hide: hideColumn,
      };
    });

    const parsedRows = allRows.slice(1).map((line, id) => {
      const row = line.split(/,/);
      const parsedRow = {};
      parsedHeader.forEach((headerName, i) => {
        if (headerName === 'id') {
          parsedRow[headerName] = id;
        } else {
          parsedRow[headerName] = row[i - 1];
        }
      });

      return parsedRow;
    });

    return (
      <DataGrid
        autoHeight
        rows={parsedRows}
        columns={parsedColumns}
        pageSize={5}
        rowsPerPageOptions={[5]}
        disableSelectionOnClick
      />
    );
  };
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
        onFiles={(files) => {
          if (files.length > 1) {
            setErrMsg('Please upload just one file.');
          } else {
            const file = files[0];
            if (file.name.slice(-4) !== '.csv') {
              setErrMsg('Please upload a CSV file.');
            } else {
              name ? null : setName(file.name.slice(0, -4));
              const reader = new FileReader();
              reader.onloadend = function (event) {
                const content = event.target.result as string;
                setCSV({ name: file.name, data: content });
              };
              reader.readAsText(file);
            }
          }
        }}
        displayFile={displayFileFn}
        onReset={(_) => {
          setName('');
          setCSV(null);
        }}
      />
    </Box>
  );
};
