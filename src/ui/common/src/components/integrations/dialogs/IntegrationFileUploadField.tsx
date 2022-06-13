import { Box, Button, Input, Typography } from '@mui/material';
import { DataGrid } from '@mui/x-data-grid';
import React, { MouseEventHandler, useEffect, useRef } from 'react';

import { theme } from '../../../styles/theme/theme';
import { FileData } from '../../../utils/integrations';

export type FileEventTarget = EventTarget & { files: FileList };

type IntegrationFileUploadFieldProps = {
  label: string;
  description: string | JSX.Element;
  required: boolean;
  file: FileData;
  placeholder: string;
  onFiles: (files: FileList) => void;
  onReset: MouseEventHandler<HTMLAnchorElement>;
};

export const IntegrationFileUploadField: React.FC<
  IntegrationFileUploadFieldProps
> = ({ label, description, required, file, placeholder, onFiles, onReset }) => {
  let header, contents;
  const drop = useRef(undefined);
  const [dragging, setDragging] = React.useState(false);

  const handleDragEnter = (e) => {
    e.preventDefault();
    e.stopPropagation();

    // So drag events won't fire on children while dragging around parent.
    if (e.target === drop.current || e.target.parentElement === drop.current) {
      [...drop.current.children].map((child) => {
        child.style.pointerEvents = 'none';
      });
      setDragging(true);
    }
  };

  const handleDragLeave = (e) => {
    e.preventDefault();
    e.stopPropagation();

    // Allow pointer events again.
    if (e.target === drop.current) {
      [...drop.current.children].map((child) => {
        child.style.pointerEvents = 'auto';
      });
      setDragging(false);
    }
  };

  const handleDragOver = (e) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const handleDrop = (e) => {
    e.preventDefault();
    e.stopPropagation();

    setDragging(false);

    const { files } = e.dataTransfer;
    if (files && files.length) {
      onFiles(files);
    }
  };

  useEffect(() => {
    if (drop.current) {
      drop.current.addEventListener('dragenter', handleDragEnter);
      drop.current.addEventListener('dragleave', handleDragLeave);
      drop.current.addEventListener('dragover', handleDragOver);
      drop.current.addEventListener('drop', handleDrop);
    }
  }, []);

  if (file) {
    header = (
      <Box>
        <Typography variant="body1" component="span" sx={{ mr: 4 }}>
          <strong>{label}</strong>: {file.name}
        </Typography>
        <Button
          size="small"
          variant="outlined"
          component="span"
          onClick={onReset}
          sx={{ float: 'right' }}
        >
          Choose File
        </Button>
      </Box>
    );

    const allRows = file.data.split(/\r?\n/);
    const parsedHeader = ['id'];
    parsedHeader.push(...allRows[0].split(/,/));
    const width = 25;
    const parsedColumns = parsedHeader.map((headerName) => {
      return {
        field: headerName,
        headerName: headerName,
        width: width * headerName.length,
      };
    });
    const parsedRows = allRows.slice(1).map((line, id) => {
      const row = line.split(/,/);
      const parsedRow = { id: id };
      parsedHeader.forEach((headerName, i) => (parsedRow[headerName] = row[i]));
      return parsedRow;
    });

    const styling = {
      margin: '16px',
      height: '25vh',
      width: `max(100%-16px,${placeholder.length + 8}ch)`,
    };

    contents = (
      <Box sx={styling}>
        <DataGrid
          rows={parsedRows}
          columns={parsedColumns}
          pageSize={5}
          rowsPerPageOptions={[5]}
          disableSelectionOnClick
        />
      </Box>
    );
  } else {
    const overlay = dragging && theme.palette.gray[100];
    const styling = {
      margin: '16px',
      height: '16ch',
      width: `max(100%-16px, ${placeholder.length + 8}ch)`,
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      flexDirection: 'column',
      backgroundColor: overlay,
    };

    header = (
      <Typography variant="body1">
        <strong>{label}</strong>
      </Typography>
    );

    // drag and drop interface
    contents = (
      <Box ref={drop} border="dashed" sx={styling}>
        <Typography variant="h6">{placeholder}</Typography>
        <Button component="label" variant="outlined" sx={{ marginTop: 2 }}>
          Choose File
          <Input
            type="file"
            onChange={(e) => {
              const fileEvent: FileEventTarget = e.target as FileEventTarget;
              onFiles(fileEvent.files);
            }}
            required={required}
            sx={{ display: 'none' }}
          />
        </Button>
      </Box>
    );
  }

  return (
    <Box sx={{ my: 2 }}>
      {header}
      <Typography variant="body2">{description}</Typography>
      {contents}
    </Box>
  );
};
