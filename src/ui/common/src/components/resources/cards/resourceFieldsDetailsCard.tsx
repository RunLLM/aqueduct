import React from 'react';
import { Resource } from 'src/utils/resources';

import { AirflowCard } from './airflowCard';
import { AqueductCard } from './aqueductCard';
import { AthenaCard } from './athenaCard';
import AWSCard from './awsCard';
import { BasicDBCard } from './basicDBCard';
import { BigQueryCard } from './bigqueryCard';
import { DatabricksCard } from './databricksCard';
import ECRCard from './ecrCard';
import GARCard from './garCard';
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
  resource: Resource;

  // Controls what fields about the resource are shown. When set to true, more fields will be shown.
  detailedView: boolean;
};

export const ResourceFieldsDetailsCard: React.FC<
  ResourceFieldsDetailsCardProps
> = ({ resource, detailedView }) => {
  let serviceCard;
  switch (resource.service) {
    case 'Aqueduct':
      serviceCard = (
        <AqueductCard resource={resource} detailedView={detailedView} />
      );
      break;
    case 'Postgres':
    case 'MySQL':
    case 'Redshift':
    case 'MariaDB':
      serviceCard = (
        <BasicDBCard resource={resource} detailedView={detailedView} />
      );
      break;
    case 'Snowflake':
      serviceCard = (
        <SnowflakeCard resource={resource} detailedView={detailedView} />
      );
      break;
    case 'MongoDB':
      serviceCard = <MongoDBCard resource={resource} />;
      break;
    case 'BigQuery':
      serviceCard = <BigQueryCard resource={resource} />;
      break;
    case 'S3':
      serviceCard = <S3Card resource={resource} />;
      break;
    case 'GCS':
      serviceCard = <GCSCard resource={resource} />;
      break;
    case 'Airflow':
      serviceCard = <AirflowCard resource={resource} />;
      break;
    case 'Kubernetes':
      serviceCard = <KubernetesCard resource={resource} />;
      break;
    case 'Lambda':
      serviceCard = <LambdaCard resource={resource} />;
      break;
    case 'Databricks':
      serviceCard = (
        <DatabricksCard resource={resource} detailedView={detailedView} />
      );
      break;
    case 'Email':
      serviceCard = (
        <EmailCard resource={resource} detailedView={detailedView} />
      );
      break;
    case 'Slack':
      serviceCard = <SlackCard resource={resource} />;
      break;
    case 'Spark':
      serviceCard = <SparkCard resource={resource} />;
      break;
    case 'AWS':
      serviceCard = <AWSCard resource={resource} />;
      break;
    case 'SQLite':
      serviceCard = <SQLiteCard resource={resource} />;
      break;
    case 'Athena':
      serviceCard = (
        <AthenaCard resource={resource} detailedView={detailedView} />
      );
      break;
    case 'ECR':
      serviceCard = <ECRCard resource={resource} />;
      break;
    case 'GAR':
      serviceCard = <GARCard resource={resource} />;
      break;
    case 'Filesystem':
      serviceCard = <FilesystemCard resource={resource} />;
      break;
    default:
      serviceCard = null;
  }
  return serviceCard;
};
