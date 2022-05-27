import Box from '@mui/material/Box';
import Head from 'next/head';
import React from 'react';
import UserProfile from "../../utils/auth";
import MenuSidebar from "./menuSidebar";

export const MenuSidebarOffset = '250px';

type Props = {
    user: UserProfile;
    contentWidth?: string;
    children: React.ReactElement[]
};

export const DefaultLayout: React.FC<Props> = ({ user, contentWidth = '100%', children }) => {
    return (
        <Box
            sx={{
                width: '100%',
                height: '100%',
                position: 'fixed',
                overflow: 'auto',
            }}
        >
            <Head>
                <link rel="icon" href="/public/favicon.ico" />
            </Head>

            <Box sx={{ width: '100%', height: '100%', display: 'flex', flex: 1 }}>
                <MenuSidebar user={user} />

                {/* The margin here is fixed to be a constant (50px) more than the sidebar, which is a fixed width (200px). */}
                <Box sx={{ marginLeft: MenuSidebarOffset, width: contentWidth, marginTop: 3 }}>{children}</Box>
            </Box>
        </Box>
    );
};

export default DefaultLayout;
