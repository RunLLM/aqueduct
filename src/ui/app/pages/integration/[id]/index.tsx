import useUser from '@aqueducthq/common/src/components/hooks/useUser';
import IntegrationDetailsPage from '@aqueducthq/common/src/components/pages/integration/id/index';
import { useRouter } from 'next/router';
import React from 'react';
export { getServerSideProps } from '@aqueducthq/common/src/components/pages/getServerSideProps';

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
