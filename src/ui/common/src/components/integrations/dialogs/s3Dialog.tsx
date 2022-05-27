import Box from '@mui/material/Box';
import React, { useEffect, useState } from 'react';
import {IntegrationConfig, S3Config} from "../../../utils/integrations";
import {IntegrationTextInputField} from "./dialog";

const Placeholders: S3Config = {
    bucket: 'aqueduct',
    access_key_id: '',
    secret_access_key: '',
};

type Props = {
    setDialogConfig: (config: IntegrationConfig) => void;
};

export const S3Dialog: React.FC<Props> = ({ setDialogConfig }) => {
    const [bucket, setBucket] = useState(null);
    const [accessKeyId, setAccessKeyId] = useState(null);
    const [secretAccessKey, setSecretAccessKey] = useState(null);

    useEffect(() => {
        const config: S3Config = {
            bucket: bucket,
            access_key_id: accessKeyId,
            secret_access_key: secretAccessKey,
        };
        setDialogConfig(config);
    }, [bucket, accessKeyId, secretAccessKey]);

    return (
        <Box sx={{ mt: 2 }}>
            <IntegrationTextInputField
                spellCheck={false}
                required={true}
                label="Bucket *"
                description="The name of the S3 bucket."
                placeholder={Placeholders.bucket}
                onChange={(event) => setBucket(event.target.value)}
                value={bucket}
            />

            <IntegrationTextInputField
                spellCheck={false}
                required={true}
                label="AWS Access Key ID"
                description="The access key ID of your AWS account."
                placeholder={Placeholders.access_key_id}
                onChange={(event) => setAccessKeyId(event.target.value)}
                value={accessKeyId}
            />

            <IntegrationTextInputField
                spellCheck={false}
                required={true}
                label="AWS Secret Access Key"
                description="The secret access key of your AWS account."
                placeholder={Placeholders.secret_access_key}
                onChange={(event) => setSecretAccessKey(event.target.value)}
                value={secretAccessKey}
            />
        </Box>
    );
};
