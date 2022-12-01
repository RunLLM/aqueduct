import Box from '@mui/material/Box';
import React from 'react';

import { SparkConfig } from '../../../utils/integrations';
import { readOnlyFieldDisableReason, readOnlyFieldWarning } from './constants';
import { IntegrationTextInputField } from './IntegrationTextInputField';

const Placeholders: SparkConfig = {
    app_name: 'app_name',
    driver_host: 'driver_host',
    master: 'master'
};

type Props = {
    onUpdateField: (field: keyof SparkConfig, value: string) => void;
    value?: SparkConfig;
    editMode: boolean;
};

export const SparkDialog: React.FC<Props> = ({
    onUpdateField,
    value,
    editMode,
}) => {
    return (
        <Box sx={{ mt: 2 }}>
            <IntegrationTextInputField
                label={'App Name*'}
                description={'The name of your Spark application'}
                spellCheck={false}
                required={true}
                placeholder={Placeholders.app_name}
                onChange={(event) => onUpdateField('app_name', event.target.value)}
                value={value?.app_name ?? null}
            />

            <IntegrationTextInputField
                label={'Driver Host*'}
                description={'The driver host to connect to.'}
                spellCheck={false}
                required={true}
                placeholder={Placeholders.driver_host}
                onChange={(event) => onUpdateField('driver_host', event.target.value)}
                value={value?.driver_host ?? null}
                disabled={editMode}
                warning={editMode ? undefined : readOnlyFieldWarning}
                disableReason={editMode ? readOnlyFieldDisableReason : undefined}
            />

            <IntegrationTextInputField
                label={'Master*'}
                description={'The master to connect to.'}
                spellCheck={false}
                required={true}
                placeholder={Placeholders.master}
                onChange={(event) => onUpdateField('master', event.target.value)}
                value={value?.master ?? null}
                disabled={editMode}
                warning={editMode ? undefined : readOnlyFieldWarning}
                disableReason={editMode ? readOnlyFieldDisableReason : undefined}
            />
        </Box>
    );
};
