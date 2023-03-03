import { faTags } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import Box from '@mui/material/Box';
import Link from '@mui/material/Link';
import Typography from '@mui/material/Typography';
import React, { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Link as RouterLink } from 'react-router-dom';

import { handleGetServerConfig } from '../../../handlers/getServerConfig';
import { RootState } from '../../../stores/store';
import { DataPreviewInfo } from '../../../utils/data';
import { getPathPrefix } from '../../../utils/getPathPrefix';
import { Integration } from '../../../utils/integrations';
import ExecutionChip from '../../execution/chip';
import useUser from '../../hooks/useUser';
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
import { LoadSpecsCard } from './loadSpecCard';
import { MariaDbCard } from './mariadbCard';
import { MongoDBCard } from './mongoDbCard';
import { MySqlCard } from './mysqlCard';
import { PostgresCard } from './postgresCard';
import { RedshiftCard } from './redshiftCard';
import { S3Card } from './s3Card';
import { SlackCard } from './slackCard';
import { SnowflakeCard } from './snowflakeCard';
import { SparkCard } from './sparkCard';

type DataProps = {
  dataPreviewInfo: DataPreviewInfo;
};

export const DataCard: React.FC<DataProps> = ({ dataPreviewInfo }) => {
  const dataPreviewInfoVersions = Object.entries(dataPreviewInfo.versions);
  if (dataPreviewInfoVersions.length > 0) {
    let [latestDagResultId, latestVersion] = dataPreviewInfoVersions[0];
    // Find the latest version
    // note: could also sort the array and get things that way.
    dataPreviewInfoVersions.forEach(([dagResultId, version]) => {
      if (version.timestamp > latestVersion.timestamp) {
        latestDagResultId = dagResultId;
        latestVersion = version;
      }
    });

    const workflowId = dataPreviewInfo.workflow_id;
    return (
      <Link
        underline="none"
        color="inherit"
        to={`${getPathPrefix()}/workflow/${workflowId}/result/${latestDagResultId}/artifact/${
          dataPreviewInfo.artifact_id
        }`}
        component={RouterLink}
      >
        <Box sx={{ display: 'flex', flexDirection: 'column' }}>
          <Box sx={{ display: 'flex', alignItems: 'center' }}>
            <Box sx={{ flex: 1 }}>
              <Typography
                variant="h6"
                component="div"
                sx={{
                  fontFamily: 'Monospace',
                  '&:hover': { textDecoration: 'underline' },
                }}
              >
                {dataPreviewInfo.artifact_name}
              </Typography>
            </Box>
            <Box marginLeft={1}>
              <ExecutionChip status={latestVersion.status} />
            </Box>
          </Box>
          <Box sx={{ fontSize: 1, my: 1 }}>
            <Typography variant="body2">
              <strong>Workflow:</strong> {dataPreviewInfo.workflow_name}
            </Typography>
            <Typography variant="body2">
              <strong>Last Updated:</strong>{' '}
              {new Date(latestVersion.timestamp * 1000).toLocaleString()}
            </Typography>
          </Box>
          <LoadSpecsCard loadSpecs={dataPreviewInfo.load_specs} />
        </Box>
      </Link>
    );
  }
  return null;
};

type IntegrationProps = {
  integration: Integration;
};

export const IntegrationCard: React.FC<IntegrationProps> = ({
  integration,
}) => {
  const { user } = useUser();
  const dispatch = useDispatch();
  const serverConfig = useSelector(
    (state: RootState) => state.serverConfigReducer
  );

  const storageConfig = serverConfig?.config?.storageConfig;

  useEffect(() => {
    async function fetchServerConfig() {
      if (user) {
        await dispatch(handleGetServerConfig({ apiKey: user.apiKey }));
      }
    }

    fetchServerConfig();
  }, [user]);

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

  let dataStorageInfo,
    dataStorageText = null;
  if (storageConfig && storageConfig.type === 'SQLite') {
    dataStorageInfo = (
      <Box component="span">
        <FontAwesomeIcon icon={faTags} />
      </Box>
    );

    dataStorageText = (
      <Typography variant={'body2'}>
        <strong>Storage Type:</strong> {dataStorageInfo}
      </Typography>
    );
  }

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column' }}>
      <Box sx={{ display: 'flex', flexDirection: 'row' }}>
        <Box sx={{ flex: 1 }}>
          <Typography sx={{ fontFamily: 'Monospace' }} variant="h6">
            {integration.name}
          </Typography>
        </Box>
        <IntegrationLogo service={integration.service} size="small" activated />
      </Box>

      {serviceCard}

      <Typography variant="body2">
        <strong>Connected On: </strong>
        {new Date(integration.createdAt * 1000).toLocaleString()}
      </Typography>
      {dataStorageText}
    </Box>
  );
};
