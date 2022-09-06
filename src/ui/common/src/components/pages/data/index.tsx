import Box from '@mui/material/Box';
import match from 'autosuggest-highlight/match';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { getDataArtifactPreview } from '../../../reducers/dataPreview';
import { AppDispatch, RootState } from '../../../stores/store';
import UserProfile from '../../../utils/auth';
import { DataCard, dataCardName } from '../../integrations/cards/card';
import { Card } from '../../layouts/card';
import { SearchBar, filteredList } from '../../layouts/search';
import DefaultLayout from '../../layouts/default';
import { LayoutProps } from '../types';
import { Typography } from '@mui/material';

type Props = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

const DataPage: React.FC<Props> = ({ user, Layout = DefaultLayout }) => {
  const apiKey = user.apiKey;
  const dispatch: AppDispatch = useDispatch();

  useEffect(() => {
    dispatch(getDataArtifactPreview({ apiKey }));
  }, []);

  const dataCardsInfo = useSelector(
    (state: RootState) => state.dataPreviewReducer
  );

  const [filterText, setFilterText] = useState<string>('');

  const displayFilteredCards = (filteredDataCards, idx) => {
    return (
      <Box key={idx}>
        <Card>
          <DataCard dataPreviewInfo={filteredDataCards} />
        </Card>
      </Box>
    );
  }

  const noItemsMessage = (
    <Typography variant="h5">
      There are no data artifacts yet.
    </Typography>
  );
    
  const dataCards = filteredList(filterText, Object.values(dataCardsInfo.data.latest_versions), (dataCardInfo) => dataCardName(dataCardInfo), displayFilteredCards, noItemsMessage);
  
  useEffect(() => {
    document.title = 'Data | Aqueduct';
  }, []);

  const getOptionLabel = (option) => {
    // When option string is invalid, none of 'options' will be selected
    // and the component will try to directly render the input string.
    // This check prevents applying `dataCardName` to the string.
    if (typeof option === 'string') {
      return option;
    }
    return dataCardName(option);
  }

  return (
    <Layout user={user}>
      <div />
      <Box>
        <Typography variant="h2" gutterBottom component="div">
          Data
        </Typography>
        
        <SearchBar options={Object.values(dataCardsInfo.data.latest_versions)} getOptionLabel={getOptionLabel} setSearchTerm={setFilterText}/>

        <Box sx={{ my: 3, ml: 1 }}>
          <Box
            sx={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'flex-start',
              my: 1,
            }}
          >
            {dataCards}
          </Box>
        </Box>
      </Box>
    </Layout>
  );
};

export default DataPage;
