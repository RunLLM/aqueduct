import {
  faChevronDown,
  faChevronRight,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import Editor from '@monaco-editor/react';
import { TreeItem, TreeView } from '@mui/lab';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';

type Props = {
  files: Record<string, any>;
  codeHeight?: string;
  defaultFile?: string;
  defaultFileExtension?: string;
};

const MultiFileViewer: React.FC<Props> = ({
  files,
  codeHeight = '30vh',
  defaultFile = '',
  defaultFileExtension = '.py'
}) => {
  // NOTE: We're making a strong-ish assumption here that we're going to have files in a format
  // where the root dir is the name of the operator and the main function is {operator_name}.{defaultFileExtension}.
  //defaultFile ? `/${defaultFile}/${defaultFile}${defaultFileExtension}` : ''
  const [selectedFilePath, setselectedFilePathPath] = useState(
    defaultFile ? `/${defaultFile}/${defaultFile}${defaultFileExtension}` : ''
  );
  const [matches, setMatches] = useState(false);
  const [multiFileViewerTree, setMultiFileViewerTree] = useState<JSX.Element[]>([]);
  const [hasFiles, setHasFiles] = useState<boolean>(false);
  const [selected, setSelected] = useState(null);

  useEffect(() => {
    const media = window.matchMedia('(min-width: 1000px)');
    if (media.matches !== matches) {
      setMatches(media.matches);
    }
    const listener = () => setMatches(media.matches);
    window.addEventListener('resize', listener);
    return () => window.removeEventListener('resize', listener);
  }, [matches]);

  useEffect(() => {
    const currentFile = getCurrentFile();
    console.log('initial load currentFile: ', currentFile);
    setSelected(currentFile);
  }, []);

  useEffect(() => {
    //console.log('defaultFile useEffect')
    setselectedFilePathPath(defaultFile ? `/${defaultFile}/${defaultFile}${defaultFileExtension}` : '');
  }, [defaultFile]);

  useEffect(() => {
    //console.log('defaultFileExtension useEffect');
    setselectedFilePathPath(defaultFile ? `/${defaultFile}/${defaultFile}${defaultFileExtension}` : '');
  }, [defaultFileExtension]);

  const isFile = (object) => {
    const keys = Object.keys(object);
    const hasFileKeys = keys.includes("path") && keys.includes("language") && keys.includes("content");
    return hasFileKeys;

    /*
      {
        "": {},
        "aqueduct_demo query 1.sql": {
            "path": "aqueduct_demo query 1/aqueduct_demo query 1.sql",
            "language": "sql",
            "content": "SELECT * FROM customers;"
        }
      }
  */
  };

  useEffect(() => {
    const filesArePresent = files && Object.keys(files).length > 0;
    setHasFiles(filesArePresent);
  }, [files]);

  useEffect(() => {
    //console.log('defaultFileExtension, hasFiles useEffect if statement');
    if (hasFiles) {
      setMultiFileViewerTree(buildTree(files, ''));
    }
  }, [defaultFileExtension, hasFiles]);

  // // TODO: refactor me into a useEffect
  // let selected = files;

  // // TODO: Refactor me into a useEffect
  // console.log('isFile(selected): ', isFile(selected));
  // if (!isFile(selected)) {
  //   // Return the default "file"
  //   selected = files[''];
  // }

  const getCurrentFile = () => {
    let currentFile = files;
    if (hasFiles) {
      const pathList = selectedFilePath.split('/').splice(1);
      //console.log('pathList: ', pathList);

      pathList.forEach((section) => {
        if (Object.keys(selected).includes(section)) {
          //console.log('setting currentFile after pathList');
          //console.log('selected2: ', selected);
          //console.log('section: ', section);
          currentFile = selected[section];
          //console.log('selected[section]: ', currentFile);
        } else {
          //console.log('setting hasFiles to false');
          setHasFiles(false);
        }
      });
    }

    //console.log('currentFile before if check: ', currentFile);
    if (!isFile(currentFile)) {
      // Return the default "file"
      currentFile = files[''];
    }

    //console.log('currentFile before return: ', currentFile);
    return currentFile;
  }

  useEffect(() => {
    const currentFile = getCurrentFile();
    setSelected(currentFile);
  }, [files, hasFiles]) // may want to trigger this one when hasFiles changes. hope that this doesn't cause infinite loop.

  const buildTree = (currentDirectory, prefix) => {
    console.log('buildTree currentDirectory: ', currentDirectory);
    console.log('buildTree prefix: ', prefix);

    const keys = Object.keys(currentDirectory);
    console.log('buildTree keys: ', keys);
    if (keys.length > 0) {
      const files = [];
      const folders = [];
      keys.forEach((section) => {
        console.log('build tree section: ', section);
        console.log('buidl tree currentDirectory[section]:', currentDirectory[section]);
        if (isFile(currentDirectory[section])) {
          console.log('pushing section to files: ', section);
          files.push(section);
        } else {
          console.log('pushing section to folders:', section)
          folders.push(section);
        }
      });

      console.log('files before sort: ', files);
      console.log('folders before sort: ', folders);

      files.sort();
      folders.sort();

      console.log('files after sort: ', files);
      console.log('folders after sort: ', folders);

      const fileItems = files.map((section) => {
        const fullPrefix = `${prefix}/${section}`;
        return (
          <TreeItem
            key={fullPrefix}
            nodeId={fullPrefix}
            label={section}
            onClick={() => setselectedFilePathPath(fullPrefix)}
          />
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

      const fileTree = [...folderItems, ...fileItems]
      return fileTree;
    }
  };

  const options = {
    readOnly: true,
    minimap: { enabled: false },
    wordWrap: 'on' as 'on' | 'off' | 'wordWrapColumn' | 'bounded',
  };

  console.log('selected before render: ', selected);
  console.log('files before render: ', files);

  console.log('selected.name: ', selected?.name); // expect this to be undefined for now.
  console.log('selected.language: ', selected?.language);
  console.log('selected.content: ', selected?.content);

  console.log('currentFile: ', getCurrentFile());

  return (
    <Box style={{ height: codeHeight, display: 'flex' }}>
      <Box style={{ width: '200px', height: '100%' }}>
        <TreeView
          aria-label="file system navigator"
          defaultCollapseIcon={<FontAwesomeIcon icon={faChevronDown} />}
          defaultExpandIcon={<FontAwesomeIcon icon={faChevronRight} />}
          expanded={[`/${defaultFile}`]}
          selected={[selectedFilePath]}
        >
          {hasFiles ? (
            multiFileViewerTree
          ) : (
            <Typography key="no_files">No files to display.</Typography>
          )}
        </TreeView>
      </Box>

      <Box
        style={{
          width: `calc(100% - 200px)`,
          height: '100%',
        }}
      >
        {
          selected && (
            <Editor
              path={selected.name}
              language={selected.language}
              value={selected.content}
              saveViewState={true}
              options={options}
            />
          )
        }
      </Box>
    </Box>
  );
};

export default MultiFileViewer;
