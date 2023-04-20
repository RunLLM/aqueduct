import { Box, Button, Input, Typography } from '@mui/material';
import React, {
  MouseEventHandler,
  useCallback,
  useEffect,
  useRef,
} from 'react';
import { useController, useFormContext } from 'react-hook-form';

import { theme } from '../../../styles/theme/theme';
import { FileData } from '../../../utils/integrations';
import { CodeBlock } from '../../CodeBlock';
import { readCredentialsFile } from './bigqueryDialog';

export type FileEventTarget = EventTarget & { files: FileList };

type IntegrationFileUploadFieldProps = {
  name: string;
  label: string;
  description: string | JSX.Element;
  required: boolean;
  file: FileData;
  placeholder: string;
  onFiles: (files: FileList) => void;
  displayFile: null | ((file: FileData) => JSX.Element);
  onReset: MouseEventHandler<HTMLAnchorElement>;
};

export const IntegrationFileUploadField: React.FC<
  IntegrationFileUploadFieldProps
> = ({
  name,
  label,
  description,
  required,
  file,
  placeholder,
  onFiles,
  displayFile,
  onReset,
}) => {
  const { control } = useFormContext();
  const { field } = useController({ control, name, rules: { required } });

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

  const handleDrop = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();

      setDragging(false);

      const { files } = e.dataTransfer;
      if (files && files.length) {
        onFiles(files);
      }
    },
    [onFiles]
  );

  useEffect(() => {
    const current = drop?.current;

    if (current) {
      current.addEventListener('dragenter', handleDragEnter);
      current.addEventListener('dragleave', handleDragLeave);
      current.addEventListener('dragover', handleDragOver);
      current.addEventListener('drop', handleDrop);
    }

    // clean up event listeners
    return () => {
      if (current) {
        current.removeEventListener('dragenter', handleDragEnter);
        current.removeEventListener('dragleave', handleDragLeave);
        current.removeEventListener('dragover', handleDragOver);
        current.removeEventListener('drop', handleDrop);
      }
    };
  }, [handleDrop]);

  if (file) {
    // If displayFile is not null, interpret and display file with the function. Otherwise, display as text.
    header = (
      <Box>
        <Typography variant="body1" component="span" sx={{ mr: 4 }}>
          <strong>{label}</strong>
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

    const styling = {
      margin: '16px',
      maxHeight: '25vh',
      width: `max(100%-16px,${placeholder.length + 8}ch)`,
    };

    contents = (
      <Box sx={styling}>
        {displayFile ? (
          displayFile(file)
        ) : (
          <CodeBlock language="plaintext">{file.data}</CodeBlock>
        )}
      </Box>
    );
  } else {
    // Upload file
    const overlay = dragging && theme.palette.gray[100];
    const styling = {
      margin: '16px',
      padding: '16px',
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
            ref={field.ref}
            type="file"
            onChange={(e) => {
              const fileEvent: FileEventTarget = e.target as FileEventTarget;
              onFiles(fileEvent.files);
              const file = fileEvent.files[0];
              readCredentialsFile(file, field.onChange);
            }}
            required={required}
            sx={{ display: 'none' }}
            onBlur={field.onBlur}
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
