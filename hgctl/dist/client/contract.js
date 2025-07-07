"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.ContractClient = void 0;
exports.createContractClient = createContractClient;
const ethers_1 = require("ethers");
const path_1 = __importDefault(require("path"));
const fs_1 = require("fs");
// Load contract ABIs from the contracts package
const CONTRACTS_PATH = path_1.default.join(__dirname, '../../../contracts/out');
class ContractClient {
    provider;
    releaseManager;
    logger;
    constructor(context, logger) {
        this.logger = logger;
        if (!context.rpcUrl) {
            throw new Error('No RPC URL configured');
        }
        this.provider = new ethers_1.ethers.JsonRpcProvider(context.rpcUrl);
        if (!context.releaseManagerAddress) {
            throw new Error('No ReleaseManager address configured');
        }
        // Load ReleaseManager ABI from compiled contracts
        const releaseManagerPath = path_1.default.join(CONTRACTS_PATH, 'ReleaseManager.sol/ReleaseManager.json');
        let releaseManagerArtifact;
        try {
            releaseManagerArtifact = JSON.parse((0, fs_1.readFileSync)(releaseManagerPath, 'utf-8'));
        }
        catch (error) {
            throw new Error(`Failed to load ReleaseManager ABI: ${error}`);
        }
        this.releaseManager = new ethers_1.ethers.Contract(context.releaseManagerAddress, releaseManagerArtifact.abi, this.provider);
    }
    async getReleases(avsAddress, operatorSetId, limit = 10) {
        const operatorSet = {
            avs: avsAddress,
            id: operatorSetId
        };
        try {
            const totalReleases = await this.releaseManager.getTotalReleases(operatorSet);
            if (totalReleases === 0n) {
                return [];
            }
            const start = totalReleases > BigInt(limit)
                ? totalReleases - BigInt(limit)
                : 0n;
            const releases = [];
            for (let i = start; i < totalReleases; i++) {
                try {
                    const release = await this.releaseManager.getRelease(operatorSet, i);
                    releases.push({
                        id: i.toString(),
                        artifacts: release.artifacts.map((a) => ({
                            digest: a.digest,
                            registryUrl: a.registryUrl
                        })),
                        upgradeByTime: Number(release.upgradeByTime)
                    });
                }
                catch (error) {
                    this.logger.warn(`Failed to fetch release ${i}: ${error}`);
                }
            }
            return releases;
        }
        catch (error) {
            this.logger.error(`Failed to get releases: ${error}`);
            throw error;
        }
    }
    async getLatestRelease(avsAddress, operatorSetId) {
        const operatorSet = {
            avs: avsAddress,
            id: operatorSetId
        };
        try {
            const [releaseId, release] = await this.releaseManager.getLatestRelease(operatorSet);
            if (releaseId === 0n && !release.artifacts?.length) {
                return null;
            }
            return {
                id: releaseId.toString(),
                artifacts: release.artifacts.map((a) => ({
                    digest: a.digest,
                    registryUrl: a.registryUrl
                })),
                upgradeByTime: Number(release.upgradeByTime)
            };
        }
        catch (error) {
            this.logger.error(`Failed to get latest release: ${error}`);
            throw error;
        }
    }
    async getCurrentRelease(avsAddress, operatorSetId) {
        const operatorSet = {
            avs: avsAddress,
            id: operatorSetId
        };
        try {
            const [releaseId, release] = await this.releaseManager.getCurrentRelease(operatorSet);
            if (releaseId === 0n && !release.artifacts?.length) {
                return null;
            }
            return {
                id: releaseId.toString(),
                artifacts: release.artifacts.map((a) => ({
                    digest: a.digest,
                    registryUrl: a.registryUrl
                })),
                upgradeByTime: Number(release.upgradeByTime)
            };
        }
        catch (error) {
            this.logger.error(`Failed to get current release: ${error}`);
            throw error;
        }
    }
}
exports.ContractClient = ContractClient;
function createContractClient(context, logger) {
    return new ContractClient(context, logger);
}
//# sourceMappingURL=contract.js.map