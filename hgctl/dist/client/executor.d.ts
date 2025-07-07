import { Logger } from '../logger';
import { Context } from '../config/context';
export interface Performer {
    performer_id: string;
    avs_address: string;
    status: string;
    application_healthy: boolean;
    resource_healthy: boolean;
    artifact_digest?: string;
}
export interface DeployArtifactRequest {
    avsAddress: string;
    digest: string;
    registryUrl: string;
}
export interface DeployArtifactResponse {
    success: boolean;
    message?: string;
    performerId?: string;
}
export declare class ExecutorClient {
    private client;
    private logger;
    constructor(address: string, logger: Logger);
    listPerformers(avsAddress?: string): Promise<Performer[]>;
    deployArtifact(request: DeployArtifactRequest): Promise<DeployArtifactResponse>;
    removePerformer(performerId: string): Promise<void>;
    private translateError;
    close(): void;
}
export declare function createExecutorClient(context: Context, logger: Logger): ExecutorClient;
//# sourceMappingURL=executor.d.ts.map