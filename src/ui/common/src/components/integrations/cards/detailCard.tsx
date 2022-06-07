import { BigQueryCard } from '@components/integrations/cards/bigqueryCard';
import { GithubCard } from '@components/integrations/cards/githubCard';
import { GoogleSheetsCard } from '@components/integrations/cards/googlesheetsCard';
import { MariaDbCard } from '@components/integrations/cards/mariadbCard';
import { MySqlCard } from '@components/integrations/cards/mysqlCard';
import { PostgresCard } from '@components/integrations/cards/postgresCard';
import { RedshiftCard } from '@components/integrations/cards/redshiftCard';
import { S3Card } from '@components/integrations/cards/s3Card';
import { SalesforceCard } from '@components/integrations/cards/salesforceCard';
import { SnowflakeCard } from '@components/integrations/cards/snowflakeCard';
import { SpiralDemoCard } from '@components/integrations/cards/spiralDemoCard';
import { SqlServerCard } from '@components/integrations/cards/sqlserverCard';
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import { Integration, SupportedIntegrations } from '@utils/integrations';
import React from 'react';

type DetailIntegrationCardProps = {
    integration: Integration;
};

export const DetailIntegrationCard: React.FC<DetailIntegrationCardProps> = ({ integration }) => {
    let serviceCard;
    switch (integration.service) {
        case 'Postgres':
            serviceCard = <PostgresCard integration={integration} />;
            break;
        case 'Snowflake':
            serviceCard = <SnowflakeCard integration={integration} />;
            break;
        case 'Spiral Demo':
            serviceCard = <SpiralDemoCard integration={integration} />;
            break;
        case 'MySQL':
            serviceCard = <MySqlCard integration={integration} />;
            break;
        case 'Redshift':
            serviceCard = <RedshiftCard integration={integration} />;
            break;
        case 'MariaDB':
            serviceCard = <MariaDbCard integration={integration} />;
            break;
        case 'SQL Server':
            serviceCard = <SqlServerCard integration={integration} />;
            break;
        case 'BigQuery':
            serviceCard = <BigQueryCard integration={integration} />;
            break;
        case 'Google Sheets':
            serviceCard = <GoogleSheetsCard integration={integration} />;
            break;
        case 'Salesforce':
            serviceCard = <SalesforceCard integration={integration} />;
        case 'S3':
            serviceCard = <S3Card integration={integration} />;
            break;
        case 'Github':
            serviceCard = <GithubCard />;
            break;
        default:
            serviceCard = null;
    }

    return (
        <Box sx={{ display: 'flex', flexDirection: 'column', width: '900px', mt: 2, mb: 2 }}>
            <Box sx={{ display: 'flex', flexDirection: 'row' }}>
                <img height="45px" src={SupportedIntegrations[integration.service].logo} />
                <Box sx={{ ml: 3 }}>
                    <Typography sx={{ fontFamily: 'Monospace' }} variant="h4">
                        {integration.name}
                    </Typography>

                    <Typography variant="body1">
                        <strong>Connected On: </strong>
                        {new Date(integration.createdAt * 1000).toLocaleString()}
                    </Typography>

                    {serviceCard}
                </Box>
            </Box>
        </Box>
    );
};
