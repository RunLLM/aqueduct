import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { BreadcrumbLink } from '../../../components/layouts/NavBar';
import { getDataArtifactPreview } from '../../../reducers/dataPreview';
import { handleLoadIntegrations } from '../../../reducers/integrations';
import { AppDispatch, RootState } from '../../../stores/store';
import UserProfile from '../../../utils/auth';
import { DataPreviewInfo } from '../../../utils/data';
import { DataCard } from '../../integrations/cards/card';
import { Card, CardPadding } from '../../layouts/card';
import DefaultLayout from '../../layouts/default';
import { filteredList, SearchBar } from '../../Search';
import { LayoutProps } from '../types';

type Props = {
  user: UserProfile;
  Layout?: React.FC<LayoutProps>;
};

const DataPage: React.FC<Props> = ({ user, Layout = DefaultLayout }) => {
  const apiKey = user.apiKey;
  const dispatch: AppDispatch = useDispatch();

  useEffect(() => {
    dispatch(getDataArtifactPreview({ apiKey }));
    dispatch(handleLoadIntegrations({ apiKey }));
  }, [apiKey, dispatch]);

  const dataCardsInfo = useSelector(
    (state: RootState) => state.dataPreviewReducer
  );

  const [filterText, setFilterText] = useState<string>('');

  const displayFilteredCards = (filteredDataCards, idx) => {
    return (
      <Card key={idx} my={2}>
        <DataCard dataPreviewInfo={filteredDataCards} />
      </Card>
    );
  };

  const noItemsMessage = (
    <Typography variant="h5">There are no data artifacts yet.</Typography>
  );

  const dataCards = filteredList(
    filterText,
    Object.values(dataCardsInfo.data.latest_versions),
    (dataCardInfo: DataPreviewInfo) => dataCardInfo.artifact_name,
    displayFilteredCards,
    noItemsMessage
  );

  useEffect(() => {
    document.title = 'Data | Aqueduct';
  }, []);

  const getOptionLabel = (option: DataPreviewInfo) => {
    // When option string is invalid, none of 'options' will be selected
    // and the component will try to directly render the input string.
    // This check prevents applying `dataCardName` to the string.
    if (typeof option === 'string') {
      return option;
    }
    return option.artifact_name;
  };

  return (
    <Layout
      breadcrumbs={[BreadcrumbLink.HOME, BreadcrumbLink.DATA]}
      user={user}
    >
      <div />
      <Box>
        <Box paddingLeft={CardPadding}>
          {/* Aligns search bar to card text */}
          <SearchBar
            options={Object.values(dataCardsInfo.data.latest_versions)}
            getOptionLabel={getOptionLabel}
            setSearchTerm={setFilterText}
          />
        </Box>
        {dataCards}
      </Box>
    </Layout>
  );
};

export default DataPage;
