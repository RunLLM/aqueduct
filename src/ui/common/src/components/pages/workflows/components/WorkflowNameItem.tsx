import { Box, Typography } from "@mui/material";
import Status from "../../../../components/workflows/workflowStatus";
import ExecutionStatus from "../../../../utils/shared";

interface WorkflowNameItemProps {
    name: string;
    status: ExecutionStatus;
}

export const WorkflowNameItem: React.FC<WorkflowNameItemProps> = ({
    name,
    status,
}) => {
    return (
        <Box display="flex" alignItems="left" justifyContent="space-between">
            <Status status={status} />
            <Typography sx={{ justifyContent: 'right' }} variant="body1">
                {name}
            </Typography>
        </Box>
    );
};

export default WorkflowNameItem;
