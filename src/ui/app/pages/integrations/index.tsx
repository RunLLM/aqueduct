import useUser from '@aqueducthq/common/src/components/hooks/useUser';
import IntegrationsPage from '@aqueducthq/common/src/components/pages/integrations';
import { useRouter } from 'next/router';
import React from 'react';
export { getServerSideProps } from '@aqueducthq/common/src/components/pages/getServerSideProps';

const Integrations: React.FC = () => {
    const router = useRouter();
    const { user, loading, success } = useUser();

    if (loading) {
        return null;
    }

    if (!success) {
        router.push('/login');
        return null;
    }

    return <IntegrationsPage user={user} />;
};

export default Integrations;
