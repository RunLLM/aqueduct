import { TextField } from '@mui/material';
import React from 'react';
import { RetentionPolicy } from 'src/utils/workflows';

type Props = {
  retentionPolicy?: RetentionPolicy;
  setRetentionPolicy: (p?: RetentionPolicy) => void;
};

const RetentionPolicySelector: React.FC<Props> = ({
  retentionPolicy,
  setRetentionPolicy,
}) => {
  let value = '';
  let helperText: string = undefined;
  if (!retentionPolicy || retentionPolicy.k_latest_runs <= 0) {
    helperText = 'Aqueduct will store all versions of this workflow.';
  } else {
    value = retentionPolicy.k_latest_runs.toString();
  }

  return (
    <TextField
      size="small"
      label="The number of latest versions to keep. Older versions will be removed."
      fullWidth
      type="number"
      value={value}
      onChange={(e) => {
        const kLatestRuns = parseInt(e.target.value);
        if (kLatestRuns <= 0 || isNaN(kLatestRuns)) {
          // Internal representation of no retention.
          setRetentionPolicy({ k_latest_runs: -1 });
          return;
        }

        setRetentionPolicy({ k_latest_runs: kLatestRuns });
      }}
      helperText={helperText}
    />
  );
};

export default RetentionPolicySelector;
