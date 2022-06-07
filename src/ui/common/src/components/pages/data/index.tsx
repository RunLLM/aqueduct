import { faSearch } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { InputAdornment, TextField, Typography } from '@mui/material';
import { Autocomplete } from '@mui/material';
import Box from '@mui/material/Box';
import match from 'autosuggest-highlight/match';
import parse from 'autosuggest-highlight/parse';
import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { getDataArtifactPreview } from '../../../reducers/dataPreview';
import { AppDispatch, RootState } from '../../../stores/store';
import UserProfile from '../../../utils/auth';
import { DataPreviewInfo } from '../../../utils/data';
import { LoadingStatusEnum } from '../../../utils/shared';
import { DataCard, dataCardName } from '../../integrations/cards/card';
import { Card } from '../../layouts/card';
import DefaultLayout from '../../layouts/default';

type Props = {
  user: UserProfile;
};

const SearchBar = (options: DataPreviewInfo[], onChangeFn: (any) => void) => {
  return (
    <Autocomplete
      sx={{ width: 300 }}
      options={options}
      freeSolo
      getOptionLabel={dataCardName}
      renderInput={(params) => {
        params['InputProps']['startAdornment'] = (
          <InputAdornment position="start">
            <FontAwesomeIcon icon={faSearch} />
          </InputAdornment>
        );
        return (
          <TextField
            {...params}
            label="Search"
            variant="standard"
            onChange={onChangeFn}
          />
        );
      }}
      renderOption={(props, option, { inputValue }) => {
        const label = dataCardName(option);
        // Matches only matches if the inputValue matches the start of any word (separated by space.)
        // We may want to modify the functionality in the future because many workflow and artifact names
        // are also hypen or underscore separated.
        const matches = match(label, inputValue);
        const parts = parse(label, matches);
        return (
          <li {...props}>
            <div>
              {parts.map((part, index) => (
                <span
                  key={index}
                  style={{
                    fontWeight: part.highlight ? 700 : 400,
                  }}
                >
                  {part.text}
                </span>
              ))}
            </div>
          </li>
        );
      }}
    />
  );
};

const DataPage: React.FC<Props> = ({ user }) => {
  const apiKey = user.apiKey;
  const dispatch: AppDispatch = useDispatch();

  useEffect(() => {
    dispatch(getDataArtifactPreview({ apiKey }));
  }, []);

  const dataCardsInfo = useSelector(
    (state: RootState) => state.dataPreviewReducer
  );
  if (dataCardsInfo.loadingStatus.loading === LoadingStatusEnum.Failed) {
    console.log(dataCardsInfo.loadingStatus.err);
  }

  const [filterText, setFilterText] = useState<string>('');

  const dataCards = Object.values(dataCardsInfo.data.latest_versions)
    .filter((dataCardInfo) => {
      if (filterText.length > 0) {
        return match(dataCardName(dataCardInfo), filterText).length > 0;
      }
      return true;
    })
    .map((filteredDataCards, idx) => {
      return (
        <Box key={idx}>
          <Card>
            <DataCard dataPreviewInfo={filteredDataCards} />
          </Card>
        </Box>
      );
    });
  const noDataText = <Typography variant="h5">No data to display.</Typography>;

  useEffect(() => {
    document.title = "Data | Aqueduct";
  }, [])

  return (
    <DefaultLayout user={user}>
      <div />
      <Box>
        <Typography variant="h2" gutterBottom component="div">
          Data
        </Typography>
        {SearchBar(Object.values(dataCardsInfo.data.latest_versions), (e) =>
          setFilterText(e.target.value)
        )}

        <Box sx={{ my: 3, ml: 1 }}>
          <Box
            sx={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'flex-start',
              my: 1,
            }}
          >
            {dataCards.length === 0 && noDataText}
            {dataCards}
          </Box>
        </Box>
      </Box>
    </DefaultLayout>
  );
};

export default DataPage;
