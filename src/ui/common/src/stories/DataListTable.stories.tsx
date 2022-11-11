import { Box, Link } from '@mui/material';
import React from 'react';

import CheckItem from '../components/pages/workflows/components/CheckItem';
import MetricItem, {
    MetricPreview,
} from '../components/pages/workflows/components/MetricItem';
import WorkflowNameItem from '../components/pages/workflows/components/WorkflowNameItem';
import WorkflowTable, {
    WorkflowTableData,
} from '../components/tables/WorkflowTable';
import { CheckLevel } from '../utils/operators';
import ExecutionStatus from '../utils/shared';

export const DataListTable: React.FC = () => {
    const checkPreviews = [
        {
            checkId: '1',
            name: 'min_churn',
            status: ExecutionStatus.Succeeded,
            level: CheckLevel.Error,
            value: 'True',
            timestamp: new Date().toLocaleString(),
        },
        {
            checkId: '2',
            name: 'max_churn',
            status: ExecutionStatus.Failed,
            level: CheckLevel.Error,
            value: 'True',
            timestamp: new Date().toLocaleString(),
        },
        {
            checkId: '3',
            name: 'avg_churn_check',
            // TODO: Come up with coherent color scheme for all of these different status levels.
            status: ExecutionStatus.Pending,
            level: CheckLevel.Warning,
            value: null,
            timestamp: new Date().toLocaleString(),
        },
        {
            checkId: '4',
            name: 'warning_test',
            // TODO: Come up with coherent color scheme for all of these different status levels.
            status: ExecutionStatus.Succeeded,
            level: CheckLevel.Warning,
            value: 'False',
            timestamp: new Date().toLocaleString(),
        },
        {
            checkId: '5',
            name: 'canceled_test',
            // TODO: Come up with coherent color scheme for all of these different status levels.
            status: ExecutionStatus.Canceled,
            level: CheckLevel.Warning,
            value: 'False',
            timestamp: new Date().toLocaleString(),
        },
    ];

    const checkTableItem = <CheckItem checks={checkPreviews} />;

    const metricsShort: MetricPreview[] = [
        { metricId: '1', name: 'avg_churn', value: '10' },
        { metricId: '2', name: 'sentiment', value: '100.5' },
        { metricId: '3', name: 'revenue_lost', value: '$20M' },
        { metricId: '4', name: 'more_metrics', value: '$500' },
    ];

    const metricsList = <MetricItem metrics={metricsShort} />;

    interface WorkflowLinkProps {
        title: string,
        url: string,
    }

    const WorkflowLink: React.FC<WorkflowLinkProps> = ({ url, title }) => {
        return <Link href={url}>{title}</Link>;
    };

    // TODO: Change this type to something more generic.
    // Also make this change in WorkflowsTable, I think we can just use Data here if we add JSX.element to Data's union type.
    const mockData: WorkflowTableData = {
        schema: {
            fields: [
                { name: 'name', type: 'varchar' },
                { name: 'created_at', type: 'varchar' },
                { name: 'workflow', type: 'varchar' },
                { name: 'type', type: 'varchar' },
                { name: 'metrics', type: 'varchar' },
                { name: 'checks', type: 'varchar' },
            ],
            pandas_version: '1.5.1',
        },
        data: [
            {
                // WorkflowNameItem and DataNameItem should be consolidated into one component.
                name: (
                    <WorkflowNameItem
                        name="churn_model"
                        status={ExecutionStatus.Succeeded}
                    />
                ),
                created_at: '11/1/2022 2:00PM',
                workflow: <WorkflowLink title="train_churn_model" url="/workflows" />,
                type: 'sklearn.linear, Linear Regression',
                metrics: metricsList,
                checks: checkTableItem,
            },
            {
                name: (
                    <WorkflowNameItem
                        name="predict_churn_dataset"
                        status={ExecutionStatus.Running}
                    />
                ),
                created_at: '11/1/2022 2:00PM',
                workflow: <WorkflowLink title="monthly_churn_prediction" url="/workflows" />,
                type: 'pandas.DataFrame',
                metrics: metricsList,
                checks: checkTableItem,
            },
            {
                name: (
                    <WorkflowNameItem
                        name="label_classifier"
                        status={ExecutionStatus.Pending}
                    />
                ),
                created_at: '11/1/2022 2:00PM',
                workflow: <WorkflowLink title="label_classifier_workflow" url="/workflows" />,
                type: 'parquet',
                metrics: metricsList,
                checks: checkTableItem,
            },
        ],
    };

    // TODO: Rename "WorkflowTable" to something more generic.
    return (
        <Box>
            <WorkflowTable data={mockData} />
        </Box>
    );
};

export default DataListTable;
