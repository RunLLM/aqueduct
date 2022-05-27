import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';
import {DataPreviewLoadSpec} from "../../../utils/data";

type LoadSpecsFormattingProps = {
    loadSpecs: DataPreviewLoadSpec[];
};

export const LoadSpecsCard: React.FC<LoadSpecsFormattingProps> = ({ loadSpecs }) => {
    const loadSpecList = loadSpecs.map((loadSpec, idx) => (
        <span key={`$loadspec-${idx}`}>
            <strong>Saved to</strong>: {loadSpec.service}
            {Object.keys(loadSpec.parameters).map((param, idx) => (
                <li key={`${loadSpec.service}-${idx}`}>
                    <strong>{param}</strong>: {loadSpec.parameters[param]}
                </li>
            ))}
        </span>
    ));
    return (
        <Box sx={{ display: 'flex', flexDirection: 'column' }}>
            {loadSpecList.map((loadSpec, idx) => (
                <Box key={`loadSpecDetails-${idx}`}>
                    <Typography variant="body1">{loadSpec}</Typography>
                </Box>
            ))}
        </Box>
    );
};
