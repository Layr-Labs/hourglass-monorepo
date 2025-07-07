export type OutputFormat = 'table' | 'json' | 'yaml';
export declare class OutputFormatter {
    static print(data: any, format?: OutputFormat): void;
    private static printTable;
    private static isPerformerArray;
    private static isReleaseArray;
    private static isRelease;
    private static printPerformersTable;
    private static printReleasesTable;
    private static formatAddress;
    private static formatDigest;
}
//# sourceMappingURL=index.d.ts.map