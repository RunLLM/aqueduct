import { LoginPage } from '@aqueducthq/common';
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
