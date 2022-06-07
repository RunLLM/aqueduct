import useUser from '@aqueducthq/common';
import IntegrationDetailsPage from '@aqueducthq/common';
import { useRouter } from 'next/router';
import React from 'react';
export { getServerSideProps } from '@aqueducthq/common';

const IntegrationDetails: React.FC = () => {
    const { user, loading, success } = useUser();
    if (loading) {
        return null;
    }
    const router = useRouter();
    if (!success) {
        router.push('/login');
        return null;
    }

    const integrationId = router.query.id as string;

    return <IntegrationDetailsPage user={user} integrationId={integrationId} />;
};

export default IntegrationDetails;
