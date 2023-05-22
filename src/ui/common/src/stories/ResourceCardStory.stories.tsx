import { Box } from '@mui/material';
import { ComponentMeta, ComponentStory } from '@storybook/react';
import React from 'react';

import { Card } from '../components/layouts/card';
import { ResourceCard } from '../components/resources/cards/card';
import {
  AqueductComputeConfig,
  BigQueryConfig,
  DatabricksConfig,
  EmailConfig,
  GCSConfig,
  KubernetesConfig,
  LambdaConfig,
  MariaDbConfig,
  MongoDBConfig,
  MySqlConfig,
  PostgresConfig,
  RedshiftConfig,
  Resource,
  S3Config,
  SlackConfig,
  SnowflakeConfig,
} from '../utils/resources';
import ExecutionStatus, { AWSCredentialType } from '../utils/shared';

const ResourceCard: React.FC = () => {
  const resources: Resource[] = [
    {
      id: '1',
      service: 'Postgres',
      name: 'Postgres Resource',
      config: {
        host: 'aam19861.us-east-2.amazonaws.com',
        port: '5432',
        database: 'prod',
        username: 'prod-pg-aq',
        password: 'this is a password',
      } as PostgresConfig,
      createdAt: Date.now() / 1000,
      exec_state: {
        status: ExecutionStatus.Succeeded,
      },
    },
    {
      id: '2',
      service: 'Snowflake',
      name: 'Snowflake Resource',
      config: {
        account_identifier: 'baa81868',
        warehouse: 'COMPUTE_WH',
        database: 'TEST',
        schema: 'PUBLIC',
        username: 'kingxu95',
        password: 'this is a password',
        role: 'SYSADMIN',
      } as SnowflakeConfig,
      createdAt: Date.now() / 1000,
      exec_state: {
        status: ExecutionStatus.Succeeded,
      },
    },
    {
      id: '3',
      service: 'MySQL',
      name: 'MySQL Resource',
      config: {
        host: 'aam19861.us-east-2.amazonaws.com',
        port: '1234',
        database: 'prod',
        username: 'prod-mysql-aq',
        password: 'this is a password',
      } as MySqlConfig,
      createdAt: Date.now() / 1000,
      exec_state: {
        status: ExecutionStatus.Succeeded,
      },
    },
    {
      id: '4',
      service: 'Redshift',
      name: 'Redshift Resource',
      config: {
        host: 'aam19861.us-east-2.amazonaws.com',
        port: '1234',
        database: 'prod',
        username: 'prod-redshift-aq',
        password: 'this is a password',
      } as RedshiftConfig,
      createdAt: Date.now() / 1000,
      exec_state: {
        status: ExecutionStatus.Failed,
      },
    },
    {
      id: '5',
      service: 'MariaDB',
      name: 'MariaDB Resource',
      config: {
        host: 'aam19861.us-east-2.amazonaws.com',
        port: '2222',
        database: 'prod',
        username: 'prod-mariadb-aq',
        password: 'this is a password',
      } as MariaDbConfig,
      createdAt: Date.now() / 1000,
      exec_state: {
        status: ExecutionStatus.Failed,
      },
    },
    {
      id: '6',
      service: 'MongoDB',
      name: 'MongoDB Resource',
      config: {
        auth_uri: 'mongodb://localhost:27017',
        database: 'prod',
      } as MongoDBConfig,
      createdAt: Date.now() / 1000,
      exec_state: {
        status: ExecutionStatus.Failed,
      },
    },
    {
      id: '7',
      service: 'BigQuery',
      name: 'BigQuery Resource',
      config: {
        project_id: 'aam19861',
        service_account_credentials: 'These are service account credentials',
      } as BigQueryConfig,
      createdAt: Date.now() / 1000,
      exec_state: {
        status: ExecutionStatus.Running,
      },
    },
    {
      id: '8',
      service: 'S3',
      name: 'S3 Resource',
      config: {
        type: AWSCredentialType.ConfigFilePath,
        bucket: 'resource-test-bucket',
        region: 'us-east-2',
        root_dir: 'path/to/dir',
        config_file_path: '~/.aws/credentials',
        config_file_profile: 'default',
      } as S3Config,
      createdAt: Date.now() / 1000,
      exec_state: {
        status: ExecutionStatus.Running,
      },
    },
    {
      id: '9',
      service: 'S3',
      name: 'ANother S3 Resource',
      config: {
        type: AWSCredentialType.ConfigFilePath,
        bucket: 'resource-test-bucket',
        region: 'us-east-2',
        config_file_path: '~/.aws/credentials',
        config_file_profile: 'default',
      } as S3Config,
      createdAt: Date.now() / 1000,
      exec_state: {
        status: ExecutionStatus.Running,
      },
    },
    {
      id: '10',
      service: 'GCS',
      name: 'GCS Resource',
      config: {
        bucket: 'resource-test-bucket',
        service_account_credentials: 'These are service account credentials',
      } as GCSConfig,
      createdAt: Date.now() / 1000,
      exec_state: {
        status: ExecutionStatus.Running,
      },
    },
    {
      id: '11',
      service: 'Airflow',
      name: 'My Airflow Compute',
      config: {
        host: 'aam19861.us-east-2.amazonaws.com',
        username: 'prod-airflow-aq',
        password: 'this is a password',
        s3_credentials_path: '~/.aws/credentials',
        s3_credentials_profile: 'default',
      },
      createdAt: Date.now() / 1000,
      exec_state: {
        status: ExecutionStatus.Succeeded,
      },
    },
    {
      id: '12',
      service: 'Kubernetes',
      name: 'My Kubernetes Compute and long name',
      config: {
        kubeconfig_path: '~/.kube/config',
        cluster_name: 'my_cluster',
      } as KubernetesConfig,
      createdAt: Date.now() / 1000,
      exec_state: {
        status: ExecutionStatus.Succeeded,
      },
    },
    {
      id: '13',
      service: 'Lambda',
      name: 'My Lambda Compute',
      config: {
        role_arn: 'arn:aws:iam::123456789012:role/lambda-role',
        exec_state: 'this is the exec state',
      } as LambdaConfig,
      createdAt: Date.now() / 1000,
      exec_state: {
        status: ExecutionStatus.Succeeded,
      },
    },
    {
      id: '14',
      service: 'Databricks',
      name: 'My Databricks Compute',
      config: {
        workspace_url: 'https://my-workspace.cloud.databricks.com',
        access_token: 'this is the access token',
        s3_instance_profile_arn:
          'arn:aws:iam::123456789012:instance-profile/s3-role',
        instance_pool_id: 'this is the instance pool id',
      } as DatabricksConfig,
      createdAt: Date.now() / 1000,
      exec_state: {
        status: ExecutionStatus.Succeeded,
      },
    },
    {
      id: '15',
      service: 'Spark',
      name: 'My Spark Compute',
      config: {
        livy_server_url: 'https://my-livy-server.com',
      },
      createdAt: Date.now() / 1000,
      exec_state: {
        status: ExecutionStatus.Succeeded,
      },
    },
    {
      id: '16',
      service: 'Slack',
      name: 'Slack Notifications',
      config: {
        token: 'xoxb-123456789012-1234567890123-123456789012345678901234',
        channels_serialized: '["#general"]',
        level: 'warning',
        enabled: 'false',
      } as SlackConfig,
      createdAt: Date.now() / 1000,
      exec_state: {
        status: ExecutionStatus.Succeeded,
      },
    },
    {
      id: '17',
      service: 'Slack',
      name: 'Slack Enabled',
      config: {
        token: 'xoxb-123456789012-1234567890123-123456789012345678901234',
        channels_serialized: '["#general"]',
        level: 'warning',
        enabled: 'true',
      } as SlackConfig,
      createdAt: Date.now() / 1000,
      exec_state: {
        status: ExecutionStatus.Succeeded,
      },
    },

    {
      id: '18',
      service: 'Email',
      name: 'Email Notifications',
      config: {
        host: 'smtp.gmail.com',
        port: '587',
        user: 'mysender@gmail.com',
        password: 'this is a password',
        targets_serialized: '["myemail@gmail.com"]',
        level: 'warning',
        enabled: 'false',
      } as EmailConfig,
      createdAt: Date.now() / 1000,
      exec_state: {
        status: ExecutionStatus.Succeeded,
      },
    },
    {
      id: '19',
      service: 'Email',
      name: 'Email Enabled',
      config: {
        host: 'smtp.gmail.com',
        port: '587',
        user: 'mysender@gmail.com',
        password: 'this is a password',
        targets_serialized: '["myemail@gmail.com"]',
        level: 'warning',
        enabled: 'true',
      } as EmailConfig,
      createdAt: Date.now() / 1000,
      exec_state: {
        status: ExecutionStatus.Succeeded,
      },
    },
    {
      id: '20',
      service: 'SQLite',
      name: 'Aqueduct Demo',
      config: {},
      createdAt: Date.now() / 1000,
      exec_state: {
        status: ExecutionStatus.Succeeded,
      },
    },
    {
      id: '21',
      service: 'Aqueduct',
      name: 'Aqueduct Server',
      config: {
        python_version: 'Python 3.8.10',
      } as AqueductComputeConfig,
      createdAt: Date.now() / 1000,
      exec_state: {
        status: ExecutionStatus.Succeeded,
      },
    },
    {
      id: '22',
      service: 'Aqueduct',
      name: 'Aqueduct Conda',
      config: {
        conda_config_serialized:
          '{' +
          '"conda_path":"/Users/kennethxu/opt/anaconda3",' +
          '"exec_state":"{' +
          '\\"user_logs\\":null,' +
          '\\"status\\":\\"succeeded\\",' +
          '\\"failure_type\\":null,' +
          '\\"error\\":null,' +
          '\\"timestamps\\":{\\"registered_at\\":null,\\"pending_at\\":null,\\"running_at\\":\\"2023-05-09T15:51:22.674257-07:00\\",\\"finished_at\\":\\"2023-05-09T15:54:30.699374-07:00\\"}}"}',
      } as AqueductComputeConfig,
      createdAt: Date.now() / 1000,
      exec_state: {
        status: ExecutionStatus.Succeeded,
      },
    },
  ];

  // Unique messages we want to display.
  const numWorkflowsUsingMsgs = [
    'Not currently in use',
    'Used by 1 workflow',
    'Used by 2 workflows',
  ];

  // Is missing the <Link> component that encapsulates the <Card> component.
  return (
    <Box
      sx={{
        display: 'flex',
        flexWrap: 'wrap',
        alignItems: 'flex-start',
      }}
    >
      {[...resources]
        .sort((a, b) => (a.createdAt < b.createdAt ? 1 : -1))
        .map((resource, idx) => {
          return (
            <Box key={idx} sx={{ mx: 1, my: 1 }}>
              <Card>
                <ResourceCard
                  resource={resource}
                  numWorkflowsUsingMsg={
                    numWorkflowsUsingMsgs[idx % numWorkflowsUsingMsgs.length]
                  }
                />
              </Card>
            </Box>
          );
        })}
    </Box>
  );
};

const ResourceCardTemplate: ComponentStory<typeof ResourceCard> = (args) => (
  <ResourceCard {...args} />
);

export const ResourceCardStory = ResourceCardTemplate.bind({});

export default {
  title: 'Test/ResourceCard',
  component: ResourceCard,
  argTypes: {},
} as ComponentMeta<typeof ResourceCard>;
