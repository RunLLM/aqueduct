import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Snackbar from '@mui/material/Snackbar';
import Typography from '@mui/material/Typography';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useParams } from 'react-router-dom';

import { DetailIntegrationCard } from '../../../../components/integrations/cards/detailCard';
import AddTableDialog from '../../../../components/integrations/dialogs/addTableDialog';
import DeleteIntegrationDialog from '../../../../components/integrations/dialogs/deleteIntegrationDialog';
import IntegrationDialog from '../../../../components/integrations/dialogs/dialog';
import IntegrationObjectList from '../../../../components/integrations/integrationObjectList';
import OperatorsOnIntegration from '../../../../components/integrations/operatorsOnIntegration';
import DefaultLayout, { MenuSidebarOffset } from '../../../../components/layouts/default';
import {
  handleListIntegrationObjects,
  handleLoadIntegrationOperators,
  handleTestConnectIntegration,
  resetEditStatus,
  resetTestConnectStatus,
} from '../../../../reducers/integration';
import { handleLoadIntegrations } from '../../../../reducers/integrations';
import { handleFetchAllWorkflowSummaries } from '../../../../reducers/listWorkflowSummaries';
import { AppDispatch, RootState } from '../../../../stores/store';
import UserProfile from '../../../../utils/auth';
import { Integration } from '../../../../utils/integrations';
import { isFailed, isLoading, isSucceeded } from '../../../../utils/shared';
import IntegrationOptions from '../../../integrations/options';
import { LayoutProps } from '../../types';
import {
  ZipReader,
  BlobReader,
  TextWriter
} from "@zip.js/zip.js";
import { exportFunction } from '../../../../utils/operators';
import { List, ListItem, ListItemButton, ListItemText } from '@mui/material';
import MultiFileViewer from '../../../../components/MultiFileViewer';

type FunctionDetailsPageProps = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
  maxRenderSize?: number
};

// Checked with file size=313285391 and handles that smoothly once loaded. However, takes a while to load.
const FunctionDetailsPage: React.FC<FunctionDetailsPageProps> = ({
  user,
  Layout = DefaultLayout,
  maxRenderSize = 100000000,
}) => {
  const [files, setFiles] = useState({
    "": {
      path:"",
      language:"plaintext",
      content:""
    }
  });
  const params = useParams();

  const setFileHelper = (prevState, file, fileContents) => {
    let nextState = {...prevState};
    const pathList = file.filename.split("/");
    let base = prevState;
    pathList.forEach((section, i) => {
        // Create a key for each first-level subfolder
        if (!Object.keys(base).includes(section)) {
            base[section] = {}
        }
        if (!file.directory && i+1 === pathList.length) {
          // Include the file metadata
          base[section] = fileContents
        } else {
            // Go into the subfolder
            base = base[section]
        }
    });
    return nextState;
  } 

  useEffect(() => {
    async function getFilesBlob() {
        // This is the function used to retrieve the contents in the function that generates the operator's zip file.
        const blob = await exportFunction(user, params.operatorId);
        if (blob) {
          const reader = new ZipReader(new BlobReader(blob));
          const entries = await reader.getEntries();
          entries.forEach((file) => {
            let language = "plaintext";
            if (file.filename.endsWith(".py")) {
              language = "python";
            }
            if (file.uncompressedSize < maxRenderSize) {
              file.getData(new TextWriter()).then((content) => {
                setFiles((prevState) => setFileHelper(prevState, file, {
                  path:file.filename,
                  language:language,
                  content:content,  
                }));
              });
            } else {
              setFiles((prevState) => setFileHelper(prevState, file, {
                path:file.filename,
                language:"plaintext",
                content:"We do not support viewing such large files.\nPlease download this file instead.",
              }));
            }
          });
          await reader.close();
        }
    }
    getFilesBlob();
  }, []);
  return (
    <Layout user={user} layoutType="workspace">
      <MultiFileViewer files={files} />
    </Layout>
  );
};

export default FunctionDetailsPage;