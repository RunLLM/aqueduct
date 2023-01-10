import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';

import { inferSchema, TableRow } from '../../../utils/data';
import { CSVConfig, FileData } from '../../../utils/integrations';
import PaginatedTable from '../../tables/PaginatedTable';
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
  }, [name, csv, setDialogConfig]);

  // More sophisticated CSV parser to handle file with the deliminator values in the data itself.
  // Source: https://stackoverflow.com/questions/8493195/how-can-i-parse-a-csv-string-with-javascript-which-contains-comma-in-data
  // Return 2D array.
  const splitFinder = /,|\r?\n|"(\\"|[^"])*?"/g;
  const CSVtoArray = (text: string) => {
    let currentRow = [];
    const rowsOut = [currentRow];
    let lastIndex = (splitFinder.lastIndex = 0);

    // add text from lastIndex to before a found newline or comma
    const pushCell = (endIndex: number | null) => {
      endIndex = endIndex || text.length;
      const addMe = text.substring(lastIndex, endIndex);
      // remove quotes around the item
      currentRow.push(addMe.replace(/^"|"$/g, ''));
      lastIndex = splitFinder.lastIndex;
    };

    let regexResp;
    // for each regexp match (either comma, newline, or quoted item)
    while ((regexResp = splitFinder.exec(text))) {
      const split = regexResp[0];

      // if it's not a quote capture, add an item to the current row
      // (quote captures will be pushed by the newline or comma following)
      if (!split.startsWith(`"`)) {
        const splitStartIndex = splitFinder.lastIndex - split.length;
        pushCell(splitStartIndex);

        // then start a new row if newline
        const isNewLine = /^\r?\n$/.test(split);
        if (isNewLine) {
          rowsOut.push((currentRow = []));
        }
      }
    }
    return rowsOut;
  };

  const displayFileFn = (file: FileData) => {
    const allRows = CSVtoArray(file.data);

    const parsedHeader = allRows[0];
    const parsedRows: TableRow[] = allRows.slice(1).map((line, id) => {
      const row = line;
      const parsedRow = {};
      parsedHeader.forEach((headerName, i) => {
        parsedRow[headerName] = row[i];
      });

      return parsedRow;
    });
    const schema = inferSchema(parsedRows, 'string');

    return <Box sx={{ pb: 4}}>
      <PaginatedTable data={{ schema: schema, data: parsedRows }} />
    </Box>;
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
        onReset={() => {
          setName('');
          setCSV(null);
        }}
      />
    </Box>
  );
};
