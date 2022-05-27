import Head from 'next/head';
import React from 'react';
import UserProfile from "../../utils/auth";
import DefaultLayout from "../layouts/default";
import GettingStartedTutorial from "../cards/GettingStartedTutorial";

type HomePageProps = {
    user: UserProfile;
};

const HomePage: React.FC<HomePageProps> = ({ user }) => {
    return (
        <DefaultLayout user={user}>
            <Head>
                <title>Home | Aqueduct</title>
            </Head>
            <GettingStartedTutorial user={user} />
        </DefaultLayout>
    );
};

export default HomePage;
