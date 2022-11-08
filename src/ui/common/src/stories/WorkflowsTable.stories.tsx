import React, { useState } from 'react';
import { Box, Typography } from '@mui/material';
import WorkflowTable, { WorkflowTableData } from '../components/tables/WorkflowTable';
import ExecutionStatus from '../utils/shared';
import Status from '../components/workflows/workflowStatus';
import { SupportedIntegrations } from '../utils/integrations';

export const WorkflowsTable: React.FC = () => {
    interface EngineItemProps {
        engineName: string;
        engineIconUrl: string
    };

    const EngineItem: React.FC<EngineItemProps> = ({ engineName, engineIconUrl }) => {
        return (
            <Box display="flex" alignItems="left" justifyContent="left">
                <img src={engineIconUrl} style={{ marginTop: '4px', marginRight: '8px' }} width="16px" height="16px" />
                <Typography variant="body1">{engineName}</Typography>
            </Box>
        );
    }

    interface WorkflowNameItemProps {
        name: string,
        status: ExecutionStatus
    }

    const WorkflowNameItem: React.FC<WorkflowNameItemProps> = ({ name, status }) => {
        return (
            <Box display="flex" alignItems="left" justifyContent="space-between">
                <Status status={status} />
                <Typography sx={{ justifyContent: 'right' }} variant="body1">{name}</Typography>
            </Box>
        );
    };

    interface MetricPreview {
        // used to fetch additional metrics and information to be shown in table.
        // TODO: Consider showing other metric related meta data here.
        metricId: string;
        name: string;
        value: string;
    }

    interface MetricItemProps {
        metrics: MetricPreview[];
    }

    const MetricItem: React.FC<MetricItemProps> = ({ metrics }) => {
        const [expanded, setExpanded] = useState(false);
        let metricList = [];
        let metricsToShow = metrics.length;
        if (!expanded && metrics.length > 3) {
            metricsToShow = 3;
        }

        for (let i = 0; i < metricsToShow; i++) {
            metricList.push(
                <Box display="flex" key={metrics[i].metricId}>
                    <Typography variant="body1">{metrics[i].name}</Typography>
                    <Typography variant="body1">{metrics[i].value}</Typography>
                </Box>
            );
        }

        const toggleExpanded = () => {
            setExpanded(!expanded);
        };

        const showLess = <Box><Typography variant='body1' onClick={toggleExpanded}>Show Less ...</Typography></Box>;
        const showMore = <Box><Typography variant='body1' onClick={toggleExpanded}>Show More ...</Typography></Box>;

        return (
            <Box>
                {metricList}
                {expanded ? showLess : showMore}
            </Box>
        );
    }

    const metricsShort: MetricPreview[] = [
        { metricId: '1', name: 'avg_churn', value: '10' },
        { metricId: '2', name: 'sentiment', value: '100.5' },
        { metricId: '3', name: 'revenue_lost', value: '$20M' },
        { metricId: '4', name: 'more_metrics', value: '$500' }
    ];

    const metricsList = <MetricItem metrics={metricsShort} />;

    const airflowEngine = <EngineItem engineName="airflow" engineIconUrl={SupportedIntegrations['Airflow'].logo} />;
    const lambdaEngine = <EngineItem engineName="lambda" engineIconUrl={SupportedIntegrations['Lambda'].logo} />;
    const kubernetesEngine = <EngineItem engineName="kubernetes" engineIconUrl={SupportedIntegrations['Kubernetes'].logo} />;

    const mockData: WorkflowTableData = {
        schema: {
            fields: [
                { name: 'name', type: 'varchar' },
                { name: 'last_run', type: 'varchar' },
                { name: 'engine', type: 'varchar' },
                { name: 'metrics', type: 'varchar' },
                { name: 'checks', type: 'varchar' }
            ],
            pandas_version: '1.5.1'
        },
        data: [
            {
                name: <WorkflowNameItem name="churn" status={ExecutionStatus.Succeeded} />,
                last_run: '11/1/2022 2:00PM',
                engine: airflowEngine,
                metrics: metricsList,
                checks: 'min_churn, max_churn'
            },
            {
                name: <WorkflowNameItem name="wine_ratings" status={ExecutionStatus.Failed} />,
                last_run: '11/1/2022 2:00PM',
                engine: lambdaEngine,
                metrics: metricsList,
                checks: 'min_churn, max_churn'
            },
            {
                name: <WorkflowNameItem name="diabetes_classifier" status={ExecutionStatus.Pending} />,
                last_run: '11/1/2022 2:00PM',
                engine: kubernetesEngine,
                metrics: metricsList,
                checks: 'min_churn, max_churn'
            },
            {
                name: <WorkflowNameItem name="mpg_regressor" status={ExecutionStatus.Canceled} />,
                last_run: '11/1/2022 2:00PM',
                engine: lambdaEngine,
                metrics: metricsList,
                checks: 'min_churn, max_churn'
            },
            {
                name: <WorkflowNameItem name="house_price_prediction" status={ExecutionStatus.Registered} />,
                last_run: '11/1/2022 2:00PM',
                engine: kubernetesEngine,
                metrics: metricsList,
                checks: 'min_churn, max_churn'
            },
        ]
    };

    return (
        <Box>
            <WorkflowTable data={mockData} />
        </Box>
    )
};

export default WorkflowsTable;
