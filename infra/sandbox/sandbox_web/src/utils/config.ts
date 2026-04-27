/**
 * Runtime configuration loader
 *
 * IMPORTANT: This file uses runtime URL detection to support both:
 * - Local development (localhost)
 * - LAN access via IP address (e.g., http://192.168.x.x:1101)
 * - Production deployments with dynamic hostnames
 *
 * The API URL is dynamically inferred from browser location, NOT from
 * build-time environment variables. This ensures the frontend works
 * correctly regardless of how the user accesses it.
 */

/**
 * Get API base URL based on current browser location.
 *
 * This function dynamically constructs the API URL using the actual
 * hostname/protocol from the browser's location, ensuring requests
 * go to the correct server (not hardcoded localhost or control-plane).
 *
 * @returns API base URL (e.g., "http://192.168.1.100:8000")
 */
export function getApiBaseUrl(): string {
  if (typeof window !== 'undefined') {
    const protocol = window.location.protocol;
    const hostname = window.location.hostname;
    // Use the actual hostname the user is accessing, not a build-time env var
    return `${protocol}//${hostname}:8000`;
  }
  // Fallback for SSR or non-browser environments
  return 'http://localhost:8000';
}

/**
 * Debug helper to check current API URL configuration.
 * Call from browser console: window.__getApiBaseUrl?.()
 */
export function registerDebugHelpers() {
  if (typeof window !== 'undefined') {
    (window as any).__getApiBaseUrl = getApiBaseUrl;
  }
}
