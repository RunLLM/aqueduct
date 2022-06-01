import LoginPage from '@aqueducthq/common/src/components/pages/LoginPage';
import React from 'react';

const Login: React.FC = () => {
    return <LoginPage />;
};

export async function getServerSideProps() {
    return {
        props: {},
    };
}

export default Login;
