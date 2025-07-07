"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.ExecutorClient = void 0;
exports.createExecutorClient = createExecutorClient;
const grpc = __importStar(require("@grpc/grpc-js"));
const protoLoader = __importStar(require("@grpc/proto-loader"));
const path_1 = __importDefault(require("path"));
// Load proto definitions from the ponos package
const PROTO_PATH = path_1.default.join(__dirname, '../../../ponos/protos/eigenlayer/hourglass/v1/executor/executor.proto');
const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
    keepCase: true,
    longs: String,
    enums: String,
    defaults: true,
    oneofs: true
});
const proto = grpc.loadPackageDefinition(packageDefinition);
class ExecutorClient {
    client;
    logger;
    constructor(address, logger) {
        this.logger = logger;
        // Use insecure credentials for local development
        const credentials = address.includes('localhost') || address.includes('executor:')
            ? grpc.credentials.createInsecure()
            : grpc.credentials.createSsl();
        this.client = new proto.eigenlayer.hourglass.v1.ExecutorService(address, credentials);
    }
    async listPerformers(avsAddress) {
        return new Promise((resolve, reject) => {
            const deadline = new Date();
            deadline.setSeconds(deadline.getSeconds() + 30);
            this.client.ListPerformers({ avs_address: avsAddress || '' }, { deadline }, (error, response) => {
                if (error) {
                    reject(this.translateError(error, 'list performers'));
                }
                else {
                    resolve(response.performers || []);
                }
            });
        });
    }
    async deployArtifact(request) {
        this.logger.info(`Deploying artifact for AVS ${request.avsAddress}`);
        return new Promise((resolve, reject) => {
            const deadline = new Date();
            deadline.setSeconds(deadline.getSeconds() + 300); // 5 minutes
            this.client.DeployArtifact({
                avs_address: request.avsAddress,
                digest: request.digest,
                registry_url: request.registryUrl
            }, { deadline }, (error, response) => {
                if (error) {
                    reject(this.translateError(error, 'deploy artifact'));
                }
                else {
                    resolve({
                        success: response.success,
                        message: response.message,
                        performerId: response.performer_id
                    });
                }
            });
        });
    }
    async removePerformer(performerId) {
        this.logger.info(`Removing performer ${performerId}`);
        return new Promise((resolve, reject) => {
            const deadline = new Date();
            deadline.setSeconds(deadline.getSeconds() + 60);
            this.client.RemovePerformer({ performer_id: performerId }, { deadline }, (error, response) => {
                if (error) {
                    reject(this.translateError(error, 'remove performer'));
                }
                else if (!response.success) {
                    reject(new Error(`Failed to remove performer: ${response.message}`));
                }
                else {
                    resolve();
                }
            });
        });
    }
    translateError(error, operation) {
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
    close() {
        grpc.closeClient(this.client);
    }
}
exports.ExecutorClient = ExecutorClient;
function createExecutorClient(context, logger) {
    const address = context.executorAddress || 'executor:9090';
    logger.debug(`Connecting to executor at ${address}`);
    return new ExecutorClient(address, logger);
}
//# sourceMappingURL=executor.js.map