// This file should map exactly to
// src/golang/cmd/server/handler/v2/list_storage_migrations.go

import {APIKeyRequest} from "./requests/ApiKey";
import {StorageMigrationResponse} from "./responses/storageMigration";

export type storageMigrationListRequest = APIKeyRequest & {
    status?: string;
    limit?: string;
    completedSince?: string;
};

export type storageMigrationListResponse = StorageMigrationResponse[];

export const storageMigrationListQuery = (req: storageMigrationListRequest) => ({
    url: 'storage-migrations',
    headers: {
        'api-key': req.apiKey,
        'status': req.status,
        'completed-since': req.completedSince,
    },
});
