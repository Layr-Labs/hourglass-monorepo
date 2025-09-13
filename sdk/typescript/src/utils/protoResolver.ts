import * as fs from 'fs';
import path from 'path';

/**
 * Resolves the path to a protobuf file, handling both development and packaged environments
 */
export function resolveProtoPath(protoFileName: string): string {
  const possiblePaths = [
    // Development paths
    path.join(__dirname, '../../proto', protoFileName),
    path.join(__dirname, '../proto', protoFileName),
    
    // Built/dist paths
    path.join(__dirname, '../../../proto', protoFileName),
    path.join(__dirname, '../../proto', protoFileName),
    
    // Package installation paths
    path.join(__dirname, '../../node_modules/@layr-labs/hourglass-performer/proto', protoFileName),
    path.resolve(process.cwd(), 'node_modules/@layr-labs/hourglass-performer/proto', protoFileName),
    
    // Alternative package paths
    path.resolve(process.cwd(), 'proto', protoFileName),
    path.join(process.cwd(), 'node_modules/@layr-labs/hourglass-performer/dist/proto', protoFileName),
  ];

  for (const possiblePath of possiblePaths) {
    try {
      if (fs.existsSync(possiblePath)) {
        return possiblePath;
      }
    } catch (e) {
      // Continue to next path
    }
  }

  throw new Error(
    `Could not find protobuf file: ${protoFileName}. ` +
    `Make sure the proto files are included in the package. ` +
    `Searched paths: ${possiblePaths.join(', ')}`
  );
}