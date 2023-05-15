import React from 'react';
import { Integration } from 'src/utils/integrations';

import { AirflowCard } from './airflowCard';
import { AqueductCard } from './aqueductCard';
import { AthenaCard } from './athenaCard';
import AWSCard from './awsCard';
import { BasicDBCard } from './basicDBCard';
import { BigQueryCard } from './bigqueryCard';
import { DatabricksCard } from './databricksCard';
import ECRCard from './ecrCard';
import { EmailCard } from './emailCard';
import FilesystemCard from './filesystemCard';
import { GCSCard } from './gcsCard';
import { KubernetesCard } from './kubernetesCard';
import { LambdaCard } from './lambdaCard';
import { MongoDBCard } from './mongoDbCard';
import { S3Card } from './s3Card';
import { SlackCard } from './slackCard';
import { SnowflakeCard } from './snowflakeCard';
import SparkCard from './sparkCard';
import SQLiteCard from './sqliteCard';

type ResourceFieldsDetailsCardProps = {
  integration: Integration;

  // Controls what fields about the integration are shown. When set to true, more fields will be shown.
  detailedView: boolean;
};

export const ResourceFieldsDetailsCard: React.FC<
  ResourceFieldsDetailsCardProps
> = ({ integration, detailedView }) => {
  let serviceCard;
  switch (integration.service) {
    case 'Aqueduct':
      serviceCard = (
        <AqueductCard integration={integration} detailedView={detailedView} />
      );
      break;
    case 'Postgres':
    case 'MySQL':
    case 'Redshift':
    case 'MariaDB':
      serviceCard = (
        <BasicDBCard integration={integration} detailedView={detailedView} />
      );
      break;
    case 'Snowflake':
      serviceCard = (
        <SnowflakeCard integration={integration} detailedView={detailedView} />
      );
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
      serviceCard = (
        <DatabricksCard integration={integration} detailedView={detailedView} />
      );
      break;
    case 'Email':
      serviceCard = (
        <EmailCard integration={integration} detailedView={detailedView} />
      );
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
    case 'SQLite':
      serviceCard = <SQLiteCard integration={integration} />;
      break;
    case 'Athena':
      serviceCard = (
        <AthenaCard integration={integration} detailedView={detailedView} />
      );
      break;
    case 'ECR':
      serviceCard = <ECRCard integration={integration} />;
      break;
    case 'Filesystem':
      serviceCard = <FilesystemCard integration={integration} />;
      break;
    default:
      serviceCard = null;
  }
  return serviceCard;
};
