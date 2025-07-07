export interface Context {
    name?: string;
    executorAddress: string;
    avsAddress?: string;
    operatorSetId?: number;
    networkId?: number;
    rpcUrl?: string;
    releaseManagerAddress?: string;
}
export interface Config {
    currentContext: string;
    contexts: Record<string, Context>;
}
export declare function loadContext(): Promise<Config>;
export declare function saveContext(config: Config): Promise<void>;
export declare function getContext(name: string): Promise<Context | null>;
export declare function setCurrentContext(name: string): Promise<void>;
export declare function updateContext(name: string, updates: Partial<Context>): Promise<void>;
export declare function deleteContext(name: string): Promise<void>;
export declare function listContexts(): Promise<{
    name: string;
    current: boolean;
}[]>;
//# sourceMappingURL=context.d.ts.map