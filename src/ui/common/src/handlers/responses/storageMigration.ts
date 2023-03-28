import {ExecState} from "../../utils/shared";

export type StorageMigrationResponse = {
    id: string;
    dest_integration_id: string;
    execution_state: ExecState;
    current: boolean;
}