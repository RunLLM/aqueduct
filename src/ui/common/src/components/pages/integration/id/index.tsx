import { Alert, Autocomplete, TextField, Typography } from '@mui/material';
import Box from '@mui/material/Box';
import CircularProgress from '@mui/material/CircularProgress';
import { DataGrid } from '@mui/x-data-grid';
import React, { SyntheticEvent, useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useParams } from 'react-router-dom';

import { DetailIntegrationCard } from '../../../../components/integrations/cards/detailCard';
import DefaultLayout from '../../../../components/layouts/default';
import { handleLoadIntegrations } from '../../../../reducers/integrations';
import {
  handleLoadIntegrationTable,
  tableKeyFn,
} from '../../../../reducers/integrationTableData';
import { handleLoadIntegrationTables } from '../../../../reducers/integrationTables';
import { AppDispatch, RootState } from '../../../../stores/store';
import UserProfile from '../../../../utils/auth';
import { Integration } from '../../../../utils/integrations';
import ExecutionStatus from '../../../../utils/shared';

type IntegrationDetailsPageProps = {
  user: UserProfile;
};

const IntegrationDetailsPage: React.FC<IntegrationDetailsPageProps> = ({
  user,
}) => {
  const dispatch: AppDispatch = useDispatch();
  const integrationId: string = useParams().id;
  const [table, setTable] = useState<string>('');

  // Using the ListIntegrationsRoute.
  // ENG-1036: We should create a route where we can pass in the integrationId and get the associated metadata and switch to using that.
  useEffect(() => {
    dispatch(handleLoadIntegrations({ apiKey: user.apiKey }));
    dispatch(
      handleLoadIntegrationTables({
        apiKey: user.apiKey,
        integrationId: integrationId,
      })
    );
  }, []);

  const integrations = useSelector(
    (state: RootState) => state.integrationsReducer.integrations
  );
  const integrationTables = useSelector(
    (state: RootState) => state.integrationTablesReducer.integrationTables
  );
  const tableListStatus = useSelector(
    (state: RootState) => state.integrationTablesReducer.thunkState
  );

  useEffect(() => {
    dispatch(
      handleLoadIntegrationTable({
        apiKey: user.apiKey,
        integrationId: integrationId,
        table: table,
      })
    );
  }, [table]);

  const tableKey = tableKeyFn(table);
  const [tableDataStatus, retrievedTableData] = useSelector(
    (state: RootState) => {
      let status = ExecutionStatus.Pending;
      if (state.integrationTableDataReducer.hasOwnProperty(tableKey)) {
        status = state.integrationTableDataReducer[tableKey].status;
      }
      let returnedData = null;
      if (table !== '' && status === ExecutionStatus.Succeeded) {
        const data = state.integrationTableDataReducer[tableKey].data;
        if (data !== undefined && data !== '') {
          returnedData = JSON.parse(data);
        }
      } else if (table !== '' && status === ExecutionStatus.Failed) {
        returnedData = state.integrationTableDataReducer[tableKey].err;
      }
      return [status, returnedData];
    }
  );

  // ENG-1052: We should update the route handler to give us the data in the format we want rather than needing to do post-processing in the FE side.
  const dataTable = {
    cols: [],
    rows: [],
  };
  if (
    table !== '' &&
    retrievedTableData &&
    tableDataStatus === ExecutionStatus.Succeeded
  ) {
    dataTable.cols.push({ field: '_id', hide: true });
    retrievedTableData.schema.fields.forEach((col, _) => {
      const header = `${col.name} (${col.type})`;
      dataTable.cols.push({
        field: col.name,
        headerName: header,
        minWidth: `${10 * header.length}px`,
        flex: 1,
      });
    });
    retrievedTableData.data.forEach((data, idx) => {
      data['_id'] = idx;
      dataTable.rows.push(data);
    });
  }

  let selectedIntegration = null;

  if (integrations) {
    (integrations as Integration[]).forEach((integration) => {
      if (integration.id === integrationId) {
        selectedIntegration = integration;
      }
    });
  }

  const handleChange = (
    event: SyntheticEvent<Element, Event>,
    newValue: string
  ) => {
    setTable(newValue);
  };

  const loading = tableListStatus === ExecutionStatus.Pending;
  const hasTable = table != null && table !== '';

  useEffect(() => {
    if (!selectedIntegration) {
      document.title = `Integration Details: ${selectedIntegration.name} | Aqueduct`;
    } else {
      document.title = `Integration Details | Aqueduct`;
    }
  }, []);

  if (!integrations || !selectedIntegration) {
    return null;
  }

  return (
    <DefaultLayout user={user}>
      <Box>
        <Typography variant="h2" gutterBottom component="div">
          Integration Details
        </Typography>

        <DetailIntegrationCard integration={selectedIntegration} />

        <Box sx={{ mt: 4 }}>
          <Typography variant="h4" gutterBottom component="div">
            Preview
          </Typography>
          <Autocomplete
            sx={{ width: '35ch' }}
            disablePortal
            value={table}
            onChange={handleChange}
            options={integrationTables}
            loading={loading}
            renderInput={(params) => (
              <TextField
                {...params}
                label="Base Table"
                InputProps={{
                  ...params.InputProps,
                  endAdornment: (
                    <React.Fragment>
                      {loading ? <CircularProgress size={30} /> : null}
                      {params.InputProps.endAdornment}
                    </React.Fragment>
                  ),
                }}
              />
            )}
          />

          <Box sx={{ mt: 3 }}>
            {hasTable && tableDataStatus === ExecutionStatus.Pending && (
              <Box sx={{ display: 'flex', flexDirection: 'row', mt: 3 }}>
                <CircularProgress size={30} />
                <Typography sx={{ ml: 2 }}>
                  Loading table <b>{table}</b>...
                </Typography>
              </Box>
            )}
            {hasTable && tableDataStatus === ExecutionStatus.Failed && (
              <Alert style={{ marginTop: '10px' }} severity="error">
                Table <b>{table}</b> failed to load. Try refreshing the page.{' '}
                <br />
                Error: {retrievedTableData}
              </Alert>
            )}
            {hasTable &&
              tableDataStatus === ExecutionStatus.Succeeded &&
              retrievedTableData !== '' && (
                <div style={{ height: '50vh', width: 'calc(100% - 25px)' }}>
                  <DataGrid
                    getRowId={(row) => row._id}
                    rows={dataTable.rows}
                    columns={dataTable.cols}
                    pageSize={50}
                    rowsPerPageOptions={[50]}
                    checkboxSelection
                    disableSelectionOnClick
                  />
                </div>
              )}
          </Box>
        </Box>
      </Box>
    </DefaultLayout>
  );
};

export default IntegrationDetailsPage;
