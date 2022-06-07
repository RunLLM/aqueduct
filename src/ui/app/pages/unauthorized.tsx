import Box from '@mui/material/Box';
import { useRouter } from 'next/router';
import React, { useEffect, useState } from 'react';

export const Unauthorized: React.FC = () => {
    const router = useRouter();
    const [errorMessage, setErrorMessage] = useState<string>(null);

    useEffect(() => {
        if (router.query.errorMessage) {
            setErrorMessage(Buffer.from(router.query.errorMessage as string, 'base64').toString('utf-8'));
        }
    }, [router.query]);

    return (
        <Box>
            <h1>Unauthorized</h1>
            {errorMessage && <h3>{errorMessage}</h3>}
        </Box>
    );
};

export async function getServerSideProps(context) {
    return {
        props: {},
    };
}

export default Unauthorized;
