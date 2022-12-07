import { faSearch } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, InputAdornment, TextField } from '@mui/material';
import { Autocomplete } from '@mui/material';
import match from 'autosuggest-highlight/match';
import parse from 'autosuggest-highlight/parse';
import React from 'react';

import { DataPreviewInfo } from '../../utils/data';
import { ListWorkflowSummary } from '../../utils/workflows';

type searchObjects = DataPreviewInfo | ListWorkflowSummary;

type Props = {
  options: searchObjects[];
  getOptionLabel: (v: searchObjects) => string;
  setSearchTerm: (v: string) => void;
};

export const SearchBar: React.FC<Props> = ({
  options,
  setSearchTerm,
  getOptionLabel,
}) => {
  if (options.length === 0) {
    return null;
  }
  return (
    <Autocomplete
      sx={{ width: 300 }}
      options={options}
      onInputChange={(_, val, reason) => {
        if (reason === 'clear') {
          setSearchTerm('');
          return;
        }

        setSearchTerm(val);
      }}
      freeSolo
      getOptionLabel={(option: searchObjects) => {
        if (getOptionLabel) {
          return getOptionLabel(option);
        }

        // default case, just return .name if no function provided.
        return (option as ListWorkflowSummary).name || '';
      }}
      renderInput={(params) => {
        params['InputProps']['startAdornment'] = (
          <InputAdornment position="start">
            <FontAwesomeIcon icon={faSearch} />
          </InputAdornment>
        );
        return (
          <TextField
            {...params}
            variant="standard"
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        );
      }}
      renderOption={(props, option, { inputValue }) => {
        const label = getOptionLabel(option);

        // Matches only matches if the inputValue matches the start of any word (separated by space.)
        // We may want to modify the functionality in the future because many workflow and artifact names
        // are also hypen or underscore separated.
        const matches = match(label, inputValue);
        const parts = parse(label, matches);
        return (
          <li {...props}>
            <Box>
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
            </Box>
          </li>
        );
      }}
    />
  );
};

/*
export const filteredList = (
  filterText: string,
  allItems: searchObjects[],
  matchOn: (item: searchObjects) => string,
  listItems: (item: searchObjects, idx: number) => JSX.Element,
  noItemsMessage: JSX.Element
): JSX.Element => {
  if (allItems.length === 0) {
    return noItemsMessage;
  }

  const matches = allItems
    .filter((item) => {
      if (filterText.length > 0) {
        return match(matchOn(item), filterText).length > 0;
      }
      return true;
    })
    .map(listItems);

  const noMatchesText = (
    <Typography variant="h5" marginTop="16px">
      No matches found.
    </Typography>
  );

  return matches.length === 0 ? (
    noMatchesText
  ) : (
    <Box sx={{ maxWidth: '1000px', width: '90%' }}>
      {matches.map((item, idx) => {
        return (
          <React.Fragment key={idx}>
            {item}
            {idx < matches.length - 1 && <Divider />}
          </React.Fragment>
        );
      })}
    </Box>
  );
};
*/
