// Utility types and functions for working with protobuf in TypeScript

/**
 * Convert a string to Uint8Array for protobuf bytes fields
 */
export function stringToBytes(str: string): Uint8Array {
  return new TextEncoder().encode(str);
}

/**
 * Convert Uint8Array to string from protobuf bytes fields
 */
export function bytesToString(bytes: Uint8Array): string {
  return new TextDecoder().decode(bytes);
}

/**
 * Convert a hex string to Uint8Array
 */
export function hexToBytes(hex: string): Uint8Array {
  // Remove '0x' prefix if present
  const cleanHex = hex.startsWith('0x') ? hex.slice(2) : hex;
  
  // Ensure even length
  const paddedHex = cleanHex.length % 2 === 0 ? cleanHex : '0' + cleanHex;
  
  const bytes = new Uint8Array(paddedHex.length / 2);
  for (let i = 0; i < paddedHex.length; i += 2) {
    bytes[i / 2] = parseInt(paddedHex.substr(i, 2), 16);
  }
  return bytes;
}

/**
 * Convert Uint8Array to hex string
 */
export function bytesToHex(bytes: Uint8Array): string {
  return '0x' + Array.from(bytes)
    .map(b => b.toString(16).padStart(2, '0'))
    .join('');
}

/**
 * Convert a number to Uint8Array (big-endian)
 */
export function numberToBytes(num: number): Uint8Array {
  const bytes = new Uint8Array(8);
  const view = new DataView(bytes.buffer);
  view.setBigUint64(0, BigInt(num), false); // big-endian
  return bytes;
}

/**
 * Convert Uint8Array to number (big-endian)
 */
export function bytesToNumber(bytes: Uint8Array): number {
  const view = new DataView(bytes.buffer);
  return Number(view.getBigUint64(0, false)); // big-endian
}

/**
 * Convert JSON object to Uint8Array
 */
export function jsonToBytes(obj: any): Uint8Array {
  return stringToBytes(JSON.stringify(obj));
}

/**
 * Convert Uint8Array to JSON object
 */
export function bytesToJson<T = any>(bytes: Uint8Array): T {
  return JSON.parse(bytesToString(bytes));
}