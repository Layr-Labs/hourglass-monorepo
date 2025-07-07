import { Logger } from '../logger';
import { Context } from '../config/context';
export interface Release {
    id: string;
    artifacts: Artifact[];
    upgradeByTime: number;
}
export interface Artifact {
    digest: string;
    registryUrl: string;
}
export interface OperatorSet {
    avs: string;
    id: number;
}
export declare class ContractClient {
    private provider;
    private releaseManager;
    private logger;
    constructor(context: Context, logger: Logger);
    getReleases(avsAddress: string, operatorSetId: number, limit?: number): Promise<Release[]>;
    getLatestRelease(avsAddress: string, operatorSetId: number): Promise<Release | null>;
    getCurrentRelease(avsAddress: string, operatorSetId: number): Promise<Release | null>;
}
export declare function createContractClient(context: Context, logger: Logger): ContractClient;
//# sourceMappingURL=contract.d.ts.map