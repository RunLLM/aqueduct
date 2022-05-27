import Alert from '@mui/material/Alert';
import AlertTitle from '@mui/material/AlertTitle';
import React from 'react';

export enum LogLevel {
    Info = 'info',
    Warning = 'warning',
    Error = 'error',
}

type Props = {
    logText: string;
    title?: string;
    level: LogLevel;
};

const LogBlock: React.FC<Props> = ({ logText, level, title }) => {
    return (
        <Alert severity={level}>
            {title && <AlertTitle>{title}</AlertTitle>}
            <pre>{logText}</pre>
        </Alert>
    );
};

export default LogBlock;
