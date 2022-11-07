import { Box, Typography } from '@mui/material';
import WorkflowTable, { WorkflowTableData } from '../components/tables/WorkflowTable';
import ExecutionStatus from '../utils/shared';
import Status from '../components/workflows/workflowStatus';
import { SupportedIntegrations } from '../utils/integrations';

export const WorkflowsTable: React.FC = () => {
    const computeEngines = Object.keys(SupportedIntegrations).filter((integrationKey) => {
        return SupportedIntegrations[integrationKey].category === 'compute';
    });

    for (let i = 0; i < computeEngines.length; i++) {
        console.log(computeEngines[i]);
        console.log(SupportedIntegrations[computeEngines[i]].logo);
    }

    interface EngineItemProps {
        engineName: string;
        engineIconUrl: string
    };

    const EngineItem: React.FC<EngineItemProps> = ({ engineName, engineIconUrl }) => {
        return (
            <Box display="flex">
                <img src={engineIconUrl} />
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
            <Box display="flex">
                <Status status={status} />
                <Typography variant="body1">{name}</Typography>
            </Box>
        );
    };

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
            { name: <WorkflowNameItem name="churn" status={ExecutionStatus.Succeeded} />, last_run: '11/1/2022 2:00PM', engine: engineMock, metrics: 'avg_churn, revenue lost', checks: 'min_churn, max_churn' },
            { name: <WorkflowNameItem name="wine_ratings" status={ExecutionStatus.Failed} />, last_run: '11/1/2022 2:00PM', engine: 'k8s_us_east', metrics: 'avg_churn, revenue lost', checks: 'min_churn, max_churn' },
            { name: <WorkflowNameItem name="diabetes_classifier" status={ExecutionStatus.Pending} />, last_run: '11/1/2022 2:00PM', engine: 'k8s_us_east', metrics: 'avg_churn, revenue lost', checks: 'min_churn, max_churn' },
            { name: <WorkflowNameItem name="mpg_regressor" status={ExecutionStatus.Canceled} />, last_run: '11/1/2022 2:00PM', engine: 'k8s_us_east', metrics: 'avg_churn, revenue lost', checks: 'min_churn, max_churn' },
            { name: <WorkflowNameItem name="house_price_prediction" status={ExecutionStatus.Registered} />, last_run: '11/1/2022 2:00PM', engine: 'k8s_us_east', metrics: 'avg_churn, revenue lost', checks: 'min_churn, max_churn' },
        ]
    };

    return (
        <Box>
            <WorkflowTable data={mockData} />
        </Box>
    )
};

export default WorkflowsTable;
