import { ethers } from 'ethers';
import { Logger } from '../logger';
import { Context } from '../config/context';
import path from 'path';
import { readFileSync } from 'fs';

// Load contract ABIs from local copy
const ABIS_PATH = path.join(__dirname, '../abis');

interface ContractArtifact {
  abi: any[];
}

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

export class ContractClient {
  private provider: ethers.Provider;
  private releaseManager: ethers.Contract;
  private logger: Logger;

  constructor(context: Context, logger: Logger) {
    this.logger = logger;
    
    if (!context.rpcUrl) {
      throw new Error('No RPC URL configured');
    }
    
    this.provider = new ethers.JsonRpcProvider(context.rpcUrl);
    
    if (!context.releaseManagerAddress) {
      throw new Error('No ReleaseManager address configured');
    }
    
    // Load ReleaseManager ABI from local copy
    const releaseManagerPath = path.join(ABIS_PATH, 'IReleaseManager.json');
    let releaseManagerArtifact: ContractArtifact;
    
    try {
      releaseManagerArtifact = JSON.parse(
        readFileSync(releaseManagerPath, 'utf-8')
      );
    } catch (error) {
      throw new Error(`Failed to load ReleaseManager ABI: ${error}`);
    }
    
    this.releaseManager = new ethers.Contract(
      context.releaseManagerAddress,
      releaseManagerArtifact.abi,
      this.provider
    );
  }

  async getReleases(
    avsAddress: string,
    operatorSetId: number,
    limit: number = 10
  ): Promise<Release[]> {
    const operatorSet: OperatorSet = {
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
      
      const releases: Release[] = [];
      
      for (let i = start; i < totalReleases; i++) {
        try {
          const release = await this.releaseManager.getRelease(operatorSet, i);
          releases.push({
            id: i.toString(),
            artifacts: release.artifacts.map((a: any) => ({
              digest: a.digest,
              registryUrl: a.registryUrl
            })),
            upgradeByTime: Number(release.upgradeByTime)
          });
        } catch (error) {
          this.logger.warn(`Failed to fetch release ${i}: ${error}`);
        }
      }
      
      return releases;
    } catch (error) {
      this.logger.error(`Failed to get releases: ${error}`);
      throw error;
    }
  }
}

export function createContractClient(context: Context, logger: Logger): ContractClient {
  return new ContractClient(context, logger);
}