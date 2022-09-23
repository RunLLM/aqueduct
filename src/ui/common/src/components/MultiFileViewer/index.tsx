import { ListItem, ListItemButton, ListItemText, List } from '@mui/material';
import Alert from '@mui/material/Alert';
import AlertTitle from '@mui/material/AlertTitle';
import Box from '@mui/material/Box';
import CircularProgress from '@mui/material/CircularProgress';
import Typography from '@mui/material/Typography';
import Image from 'mui-image';
import React, { useEffect, useState } from 'react';

import { ArtifactResult } from '../../reducers/workflow';
import { SerializationType } from '../../utils/artifacts';
import { ExecutionStatus, LoadingStatusEnum } from '../../utils/shared';
import { Error } from '../../utils/shared';
import { MenuSidebarOffset } from '../layouts/default';
import DataTable from '../tables/DataTable';
import LogBlock, { LogLevel } from '../text/LogBlock';
import Editor from "@monaco-editor/react";
import { TreeView, TreeItem } from '@mui/lab';

import {
  faChevronRight,
  faChevronDown,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

type Props = {
  files: Record<string, any>;
  codeHeight?: string;
};

const MultiFileViewer: React.FC<Props> = ({ files, codeHeight = '30vh' }) => {
  const [selectedFile, setSelectedFile] = useState("");
  const [matches, setMatches] = useState(false);

  useEffect(() => {
    const media = window.matchMedia("(min-width: 1000px)");
    if (media.matches !== matches) {
      setMatches(media.matches);
    }
    const listener = () => setMatches(media.matches);
    window.addEventListener("resize", listener);
    return () => window.removeEventListener("resize", listener);
  }, [matches]);

  
  const isFile = (object) => {
    return Object.keys(object).includes('language') && typeof object.language === 'string';
  }

  let hasFiles = files && Object.keys(files).length > 0;

  let selected = files;
  if (hasFiles) {
    const pathList = selectedFile.split("/").splice(1);

    pathList.forEach((section) => {
      if (Object.keys(selected).includes(section)) {
        selected = selected[section];
      } else {
        hasFiles = false;
      }
    });
  }

  if (!isFile(selected)) {
    // Return the default "file"
    selected = files[""]
  }

  const buildTree = (currentDirectory, prefix) => {
    const keys = Object.keys(currentDirectory);
    if (keys.length > 0) {
      let files = [];
      let folders = [];
      keys.forEach((section) => {
        if (isFile(currentDirectory[section])) {
          files.push(section);
        } else {
          folders.push(section);
        }
      });
      files.sort();
      folders.sort();
      const fileItems = files.map((section) => {
        const fullPrefix = `${prefix}/${section}`;
        return (
          <TreeItem key={fullPrefix} nodeId={fullPrefix} label={section} onClick={() => setSelectedFile(fullPrefix)}>
          </TreeItem>
        );
      });
      const folderItems = folders.map((section) => {
        const fullPrefix = `${prefix}/${section}`;
        return (
          <TreeItem key={fullPrefix} nodeId={fullPrefix} label={section}>
            {buildTree(currentDirectory[section], fullPrefix)}
          </TreeItem>
        );
      });
      return [...folderItems, ...fileItems];
    }
  }

  let options = {readOnly: true, minimap: { enabled: false }};
  if (matches) {
      options.minimap.enabled = true;
  }

  return (
      <Box style={{height: codeHeight}}>
        <Box style={{width: MenuSidebarOffset, float: "left", height: "100%"}}>
          <TreeView
              aria-label="file system navigator"
              defaultCollapseIcon={<FontAwesomeIcon icon={faChevronDown} />}
              defaultExpandIcon={<FontAwesomeIcon icon={faChevronRight} />}
            >
            {hasFiles? buildTree(files, "") : <Typography key="no_files">No files to display.</Typography>}
          </TreeView>
        </Box>
        <Box style={{width: `calc(100% - ${MenuSidebarOffset})`, float: "right", height: "100%"}}>
          <Editor
            path={selected.name}
            language={selected.language}
            value={selected.content}
            saveViewState={true}
            options={options}
          />
        </Box>
      </Box>
    );
};

export default MultiFileViewer;
