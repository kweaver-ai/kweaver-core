import fs from "node:fs";
import path from "node:path";
import { createHash } from "node:crypto";
import type { IncomingMessage, ServerResponse } from "node:http";
import type { OpenClawPluginApi } from "openclaw/plugin-sdk";

/** Maximum upload body size for workspace temp uploads (bytes). */
const MAX_WORKSPACE_TEMP_UPLOAD_BYTES = 32 * 1024 * 1024;

/**
 * Reads one raw request body with a hard size cap.
 *
 * @param req Incoming HTTP request.
 * @param maxBytes Maximum accepted body length in bytes.
 * @returns Concatenated request bytes.
 */
export function readRequestBodyLimited(
  req: IncomingMessage,
  maxBytes: number
): Promise<Buffer> {
  return new Promise((resolve, reject) => {
    const chunks: Buffer[] = [];
    let total = 0;

    req.on("data", (chunk: Buffer) => {
      total += chunk.length;
      if (total > maxBytes) {
        req.destroy();
        reject(new Error(`Request body exceeds ${maxBytes} bytes`));
        return;
      }
      chunks.push(chunk);
    });
    req.on("end", () => resolve(Buffer.concat(chunks)));
    req.on("error", reject);
  });
}

/**
 * Parses one multipart `file` part from request bytes.
 *
 * @param body Full request body.
 * @param contentType Request content-type header.
 * @returns Extracted filename and payload.
 */
export function parseMultipartUploadFile(
  body: Buffer,
  contentType: string
): { filename?: string; payload: Buffer } {
  const boundaryMatch = /boundary=([^\s;]+)/i.exec(contentType);
  if (boundaryMatch === null) {
    throw new Error("Multipart upload is missing boundary");
  }

  const boundary = boundaryMatch[1];
  const boundaryToken = Buffer.from(`--${boundary}`);
  let cursor = 0;

  while (true) {
    const start = body.indexOf(boundaryToken, cursor);
    if (start < 0) {
      break;
    }

    const afterBoundary = start + boundaryToken.length;
    const isFinalBoundary =
      body[afterBoundary] === 45 && body[afterBoundary + 1] === 45;
    if (isFinalBoundary) {
      break;
    }

    const partStart = afterBoundary + 2; // skip CRLF
    const nextBoundary = body.indexOf(boundaryToken, partStart);
    if (nextBoundary < 0) {
      break;
    }

    let part = body.subarray(partStart, nextBoundary);
    if (
      part.length >= 2 &&
      part[part.length - 2] === 13 &&
      part[part.length - 1] === 10
    ) {
      part = part.subarray(0, part.length - 2);
    }

    const separator = Buffer.from("\r\n\r\n");
    const headerEnd = part.indexOf(separator);
    if (headerEnd < 0) {
      cursor = nextBoundary;
      continue;
    }

    const headersText = part.subarray(0, headerEnd).toString("utf8");
    const contentDispositionLine = headersText
      .split("\r\n")
      .find((line) => line.toLowerCase().startsWith("content-disposition:"));

    if (contentDispositionLine === undefined) {
      cursor = nextBoundary;
      continue;
    }

    const nameMatch = /name="([^"]+)"/i.exec(contentDispositionLine);
    if (nameMatch?.[1] !== "file") {
      cursor = nextBoundary;
      continue;
    }

    const filenameMatch = /filename="([^"]*)"/i.exec(contentDispositionLine);
    const payload = Buffer.from(part.subarray(headerEnd + separator.length));

    return {
      filename: filenameMatch?.[1],
      payload
    };
  }

  throw new Error("Multipart field `file` is required");
}

/**
 * Sanitizes one upload filename for filesystem-safe storage.
 *
 * @param rawName User-provided filename.
 * @returns Safe basename.
 */
export function sanitizeUploadFilename(rawName?: string | null): string {
  const candidate =
    typeof rawName === "string" && rawName.trim().length > 0
      ? rawName.trim()
      : "upload.bin";
  const basename = path.basename(candidate.replaceAll("\\", "/"));
  const cleaned = basename
    .replace(/\s+/g, "_")
    .replace(/[^\p{L}\p{N}._-]+/gu, "_")
    .replace(/^_+|_+$/g, "")
    .slice(0, 128);

  if (cleaned === "" || cleaned === "." || cleaned === "..") {
    return "upload.bin";
  }

  return cleaned;
}

/**
 * Builds a hash-suffixed temp filename from one original filename.
 *
 * Example: `a.txt` -> `a_<hash>.txt`
 *
 * @param rawName Original filename hint.
 * @param payload Uploaded bytes used to compute hash.
 * @returns Hash-suffixed filename.
 */
export function buildHashedUploadFilename(
  rawName: string | null | undefined,
  payload: Buffer
): string {
  const safe = sanitizeUploadFilename(rawName);
  const initialExt = path.extname(safe);
  const ext = initialExt.length === 0 && safe.startsWith(".") ? safe : initialExt;
  const baseRaw = ext.length > 0 ? safe.slice(0, -ext.length) : safe;
  const base = baseRaw.length > 0 ? baseRaw : "upload";
  const hash = createHash("sha256").update(payload).digest("hex").slice(0, 12);

  return ext.length > 0 ? `${base}_${hash}${ext}` : `${base}_${hash}`;
}

/**
 * Normalizes one session segment used as a temporary subdirectory.
 *
 * @param rawSession Optional session id or session key.
 * @returns Safe session folder name.
 */
export function normalizeSessionSegment(rawSession?: string | null): string | undefined {
  if (typeof rawSession !== "string" || rawSession.trim().length === 0) {
    return undefined;
  }

  const trimmed = rawSession.trim();
  const source = trimmed.includes(":")
    ? trimmed.split(":").at(-1) ?? trimmed
    : trimmed;
  const normalized = source.replace(/[^a-zA-Z0-9-_]/g, "_");

  return normalized.length > 0 ? normalized : undefined;
}

/**
 * Resolves one agent workspace from runtime config.
 *
 * @param api OpenClaw plugin API.
 * @param agentId Optional agent id.
 * @returns Absolute workspace directory.
 */
async function resolveWorkspaceDir(
  api: OpenClawPluginApi,
  agentId?: string | null
): Promise<string> {
  if (agentId === undefined || agentId === null || agentId.trim().length === 0) {
    return api.resolvePath(".") || process.cwd();
  }

  const cfg = await api.runtime.config.loadConfig();
  const agentList = (cfg.agents as { list?: Array<{ id?: string; workspace?: string }> })?.list;
  const resolved = agentList?.find((entry) => entry.id === agentId.trim())?.workspace;

  if (typeof resolved !== "string" || resolved.trim().length === 0) {
    throw new Error(`Agent workspace not found for: ${agentId}`);
  }

  return resolved;
}

/**
 * Writes one upload into workspace temp area.
 *
 * @param workspaceDir Resolved workspace directory.
 * @param payload Raw file bytes.
 * @param filename Original file name hint.
 * @param session Optional session folder.
 * @returns Relative and absolute storage metadata.
 */
export async function writeWorkspaceTempUpload(
  workspaceDir: string,
  payload: Buffer,
  filename?: string | null,
  session?: string | null
): Promise<{
  name: string;
  relativePath: string;
  absolutePath: string;
  bytes: number;
}> {
  if (payload.length === 0) {
    throw new Error("Upload body is empty");
  }

  const hashedFilename = buildHashedUploadFilename(filename, payload);
  const sessionSegment = normalizeSessionSegment(session);
  const tempBaseDir = sessionSegment === undefined
    ? path.join(workspaceDir, "tmp")
    : path.join(workspaceDir, "tmp", sessionSegment);
  await fs.promises.mkdir(tempBaseDir, { recursive: true });

  const storedName = hashedFilename;
  const absolutePath = path.join(tempBaseDir, storedName);
  await fs.promises.writeFile(absolutePath, payload);

  const relativePath = sessionSegment === undefined
    ? path.join("tmp", storedName)
    : path.join("tmp", sessionSegment, storedName);

  return {
    name: storedName,
    relativePath: relativePath.replaceAll(path.sep, "/"),
    absolutePath,
    bytes: payload.length
  };
}

/**
 * Registers `POST /v1/workspace/tmp/upload` for raw file uploads.
 *
 * Query parameters:
 * - `agent` (optional): resolve and upload into that agent workspace.
 * - `session` (optional): subdirectory under `tmp/`.
 *
 * @param api OpenClaw plugin API.
 */
export function registerWorkspaceTempUpload(api: OpenClawPluginApi): void {
  api.registerHttpRoute({
    path: "/v1/workspace/tmp/upload",
    match: "prefix",
    auth: "gateway",
    handler: async (req: IncomingMessage, res: ServerResponse) => {
      const url = new URL(req.url || "", "http://localhost");
      const isExactPath =
        url.pathname === "/v1/workspace/tmp/upload" ||
        url.pathname === "/v1/workspace/tmp/upload/";

      if (!isExactPath) {
        res.statusCode = 404;
        res.setHeader("Content-Type", "application/json");
        res.end(JSON.stringify({ error: "Not found" }));
        return true;
      }

      if (req.method !== "POST") {
        res.statusCode = 405;
        res.setHeader("Content-Type", "application/json");
        res.end(
          JSON.stringify({
            error: "Method not allowed. Use POST with a raw file body."
          })
        );
        return true;
      }

      try {
        const workspaceDir = await resolveWorkspaceDir(api, url.searchParams.get("agent"));
        const body = await readRequestBodyLimited(req, MAX_WORKSPACE_TEMP_UPLOAD_BYTES);
        const contentTypeRaw = req.headers["content-type"];
        const contentType = Array.isArray(contentTypeRaw)
          ? contentTypeRaw[0] ?? ""
          : contentTypeRaw ?? "";
        const isMultipart = contentType.toLowerCase().startsWith("multipart/form-data");
        const upload = isMultipart
          ? parseMultipartUploadFile(body, contentType)
          : { payload: body, filename: undefined };
        const result = await writeWorkspaceTempUpload(
          workspaceDir,
          upload.payload,
          upload.filename,
          url.searchParams.get("session")
        );

        res.statusCode = 200;
        res.setHeader("Content-Type", "application/json");
        res.end(
          JSON.stringify({
            name: result.name,
            path: result.relativePath,
            absolutePath: result.absolutePath,
            bytes: result.bytes
          })
        );
        return true;
      } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        if (
          message.includes("exceeds") ||
          message.includes("too large")
        ) {
          res.statusCode = 413;
        } else if (
          message.includes("not found") ||
          message.includes("empty")
        ) {
          res.statusCode = 400;
        } else {
          res.statusCode = 500;
        }
        res.setHeader("Content-Type", "application/json");
        res.end(JSON.stringify({ error: message }));
        return true;
      }
    }
  });
}
