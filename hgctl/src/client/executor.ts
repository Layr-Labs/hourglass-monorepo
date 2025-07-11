import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';
import { Logger } from '../logger';
import { Context } from '../config/context';
import path from 'path';

// Load proto definitions from local copy
const PROTO_PATH = path.join(__dirname, '../protos/executor.proto');

const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true
});

const proto = grpc.loadPackageDefinition(packageDefinition) as any;

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

export class ExecutorClient {
  private client: any;
  private logger: Logger;

  constructor(address: string, logger: Logger) {
    this.logger = logger;
    
    // Use insecure credentials for local development
    const credentials = address.includes('localhost') || address.includes('executor:')
      ? grpc.credentials.createInsecure()
      : grpc.credentials.createSsl();
    
    this.client = new proto.eigenlayer.hourglass.v1.ExecutorService(
      address,
      credentials
    );
  }

  async listPerformers(avsAddress?: string): Promise<Performer[]> {
    return new Promise((resolve, reject) => {
      const deadline = new Date();
      deadline.setSeconds(deadline.getSeconds() + 30);
      
      this.client.ListPerformers(
        { avs_address: avsAddress || '' },
        { deadline },
        (error: grpc.ServiceError | null, response: any) => {
          if (error) {
            reject(this.translateError(error, 'list performers'));
          } else {
            resolve(response.performers || []);
          }
        }
      );
    });
  }

  async deployArtifact(request: DeployArtifactRequest): Promise<DeployArtifactResponse> {
    this.logger.info(`Deploying artifact for AVS ${request.avsAddress}`);
    
    return new Promise((resolve, reject) => {
      const deadline = new Date();
      deadline.setSeconds(deadline.getSeconds() + 300); // 5 minutes
      
      this.client.DeployArtifact(
        {
          avs_address: request.avsAddress,
          digest: request.digest,
          registry_url: request.registryUrl
        },
        { deadline },
        (error: grpc.ServiceError | null, response: any) => {
          if (error) {
            reject(this.translateError(error, 'deploy artifact'));
          } else {
            resolve({
              success: response.success,
              message: response.message,
              performerId: response.performer_id
            });
          }
        }
      );
    });
  }

  async removePerformer(performerId: string): Promise<void> {
    this.logger.info(`Removing performer ${performerId}`);
    
    return new Promise((resolve, reject) => {
      const deadline = new Date();
      deadline.setSeconds(deadline.getSeconds() + 60);
      
      this.client.RemovePerformer(
        { performer_id: performerId },
        { deadline },
        (error: grpc.ServiceError | null, response: any) => {
          if (error) {
            reject(this.translateError(error, 'remove performer'));
          } else if (!response.success) {
            reject(new Error(`Failed to remove performer: ${response.message}`));
          } else {
            resolve();
          }
        }
      );
    });
  }

  private translateError(error: grpc.ServiceError, operation: string): Error {
    switch (error.code) {
      case grpc.status.UNAVAILABLE:
        return new Error('Executor service unavailable - check if the executor is running');
      case grpc.status.NOT_FOUND:
        return new Error(`${operation} failed: resource not found`);
      case grpc.status.INVALID_ARGUMENT:
        return new Error(`${operation} failed: ${error.message}`);
      default:
        this.logger.debug(`gRPC error details: ${error.stack}`);
        return new Error(`${operation} failed: ${error.message}`);
    }
  }

  close(): void {
    grpc.closeClient(this.client);
  }
}

export function createExecutorClient(context: Context, logger: Logger): ExecutorClient {
  const address = context.executorAddress || 'executor:9090';
  logger.debug(`Connecting to executor at ${address}`);
  return new ExecutorClient(address, logger);
}