import Box from '@mui/material/Box';
import React from 'react';

import { Integration } from '../../../utils/integrations';
import ExecutionStatus from '../../../utils/shared';
import { StatusIndicator } from '../../workflows/workflowStatus';
import IntegrationLogo from '../logo';
import { AirflowCard } from './airflowCard';
import { AqueductDemoCard } from './aqueductDemoCard';
import { AWSCard } from './awsCard';
import { BigQueryCard } from './bigqueryCard';
import { DatabricksCard } from './databricksCard';
import { EmailCard } from './emailCard';
import { GCSCard } from './gcsCard';
import { KubernetesCard } from './kubernetesCard';
import { LambdaCard } from './lambdaCard';
import { MariaDbCard } from './mariadbCard';
import { MongoDBCard } from './mongoDbCard';
import { MySqlCard } from './mysqlCard';
import { PostgresCard } from './postgresCard';
import { RedshiftCard } from './redshiftCard';
import { S3Card } from './s3Card';
import { SlackCard } from './slackCard';
import { SnowflakeCard } from './snowflakeCard';
import { SparkCard } from './sparkCard';
import { TruncatedText } from './text';

type IntegrationProps = {
  integration: Integration;

  // Eg: "2 workflows using this integration"
  numWorkflowsUsingMsg: string;
};

const paddingRightForNumWorkflowsMsg = 8; // pixels
const paddingBottomForNumWorkflowsMsg = 4; // pixels

export const IntegrationCard: React.FC<IntegrationProps> = ({
  integration,
  numWorkflowsUsingMsg,
}) => {
  let serviceCard;
  switch (integration.service) {
    case 'Postgres':
      serviceCard = <PostgresCard integration={integration} />;
      break;
    case 'Snowflake':
      serviceCard = <SnowflakeCard integration={integration} />;
      break;
    case 'Aqueduct Demo':
      serviceCard = <AqueductDemoCard integration={integration} />;
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
    case 'MongoDB':
      serviceCard = <MongoDBCard integration={integration} />;
      break;
    case 'BigQuery':
      serviceCard = <BigQueryCard integration={integration} />;
      break;
    case 'S3':
      serviceCard = <S3Card integration={integration} />;
      break;
    case 'GCS':
      serviceCard = <GCSCard integration={integration} />;
      break;
    case 'Airflow':
      serviceCard = <AirflowCard integration={integration} />;
      break;
    case 'Kubernetes':
      serviceCard = <KubernetesCard integration={integration} />;
      break;
    case 'Lambda':
      serviceCard = <LambdaCard integration={integration} />;
      break;
    case 'Databricks':
      serviceCard = <DatabricksCard integration={integration} />;
      break;
    case 'Email':
      serviceCard = <EmailCard integration={integration} />;
      break;
    case 'Slack':
      serviceCard = <SlackCard integration={integration} />;
      break;
    case 'Spark':
      serviceCard = <SparkCard integration={integration} />;
      break;
    case 'AWS':
      serviceCard = <AWSCard integration={integration} />;
      break;
    default:
      serviceCard = null;
  }

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <Box sx={{ display: 'flex', flexDirection: 'row', alignItems: 'center' }}>
        {/*If the execution state doesn't exist, we assume the integration succeeded.*/}
        <StatusIndicator
          status={integration.exec_state?.status || ExecutionStatus.Succeeded}
          size="16px"
        />

        {/* Subtract the width of the status indicator, padding, and logo respectively. */}
        <Box
          sx={{ mx: 1, flex: 1, maxWidth: `calc(100% - 16px - 16px - 24px)` }}
        >
          <TruncatedText sx={{ fontWeight: 400 }} variant="h6">
            {integration.name}
          </TruncatedText>
        </Box>
        <IntegrationLogo service={integration.service} size="small" activated />
      </Box>

      <TruncatedText
        variant="caption"
        marginBottom={1}
        sx={{ fontWeight: 300 }}
      >
        {new Date(integration.createdAt * 1000).toLocaleString()}
      </TruncatedText>

      {serviceCard}

      <Box
        sx={{
          position: 'absolute',
          bottom: paddingBottomForNumWorkflowsMsg,
          right: paddingRightForNumWorkflowsMsg,
          textAlign: 'right',
        }}
      >
        <TruncatedText variant="caption" sx={{ fontWeight: 300 }}>
          {numWorkflowsUsingMsg}
        </TruncatedText>
      </Box>
    </Box>
  );
};
