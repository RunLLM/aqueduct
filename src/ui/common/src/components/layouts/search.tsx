import React from 'react';
import { faSearch } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { InputAdornment, TextField, Typography } from '@mui/material';
import { Autocomplete } from '@mui/material';
import parse from 'autosuggest-highlight/parse';
import match from 'autosuggest-highlight/match';

type Props = {
  options: any[];
  getOptionLabel: (v: any) => string;
  setSearchTerm: (v: string) => void;
};

export const SearchBar: React.FC<Props> = ({ options, getOptionLabel, setSearchTerm }) => {
  return (
    <Autocomplete
      sx={{ width: 300 }}
      options={options}
      onInputChange={(_, val, reason) => {
        if (reason === 'clear') {
          setSearchTerm('');
        }

        setSearchTerm(val);
      }}
      freeSolo
      getOptionLabel={getOptionLabel}
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

export const filteredList = (
  filterText: string,
  allItems: any[],
  matchOn: (item: any) => string,
  listItems: (item: any, idx: number) => JSX.Element | JSX.Element[],
  ) => {
  const matches = (allItems.filter((item) => {
    if (filterText.length > 0) {
      return match(matchOn(item), filterText).length > 0;
    }
    return true;
  })
  .map(listItems));

  const noMatchesText = <Typography variant="h5">No matches found.</Typography>;

  return matches.length === 0 ? noMatchesText : matches;
};
