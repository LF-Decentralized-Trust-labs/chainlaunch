/**
 * Decodes a base64 encoded block and converts it to JSON
 * @param base64Block - The base64 encoded block string
 * @returns The decoded block as a JSON object, or null if decoding fails
 */
export function decodeBlockToJson(base64Block: string): any | null {
  try {
    // Decode the base64 string to a UTF-8 string
    const decodedString = atob(base64Block);
    
    // Parse the decoded string as JSON
    const jsonObject = JSON.parse(decodedString);
    
    return jsonObject;
  } catch (error) {
    console.error("Failed to decode block:", error);
    return null;
  }
}
