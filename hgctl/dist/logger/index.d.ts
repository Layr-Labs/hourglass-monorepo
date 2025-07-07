export interface Logger {
    info(message: string, ...args: any[]): void;
    warn(message: string, ...args: any[]): void;
    error(message: string, ...args: any[]): void;
    debug(message: string, ...args: any[]): void;
    title(message: string, ...args: any[]): void;
}
export declare function setupLogger(verbose?: boolean): Logger;
//# sourceMappingURL=index.d.ts.map