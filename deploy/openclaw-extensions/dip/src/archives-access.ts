import fs from "node:fs";
import path from "node:path";
import type { IncomingMessage, ServerResponse } from "node:http";
import type { OpenClawPluginApi } from "openclaw/plugin-sdk";
import { ARCHIVES_MIME_MAP, formatTimestamp, sanitizeFileName } from "./archives-utils.js";

/**
 * Registers `/v1/archives` HTTP route and `after_tool_call` archive compliance hook.
 *
 * @param api OpenClaw plugin API.
 */
export function registerArchivesAccess(api: OpenClawPluginApi): void {
  api.registerHttpRoute({
    path: "/v1/archives",
    match: "prefix",
    auth: "gateway",
    handler: async (req: IncomingMessage, res: ServerResponse) => {
      api.logger.debug?.(`Incoming request to dip archives: ${req.url}`);
      try {
        let workspaceDir = api.resolvePath(".") || process.cwd();

        const urlStr = req.url || "/";
        const urlObj = new URL(urlStr, "http://localhost");
        const agentId = urlObj.searchParams.get("agent");

        if (agentId) {
          // `api.config` is a registration-time snapshot; load fresh config like skills-control.
          const cfg = await api.runtime.config.loadConfig();
          const agentsObj = cfg.agents as any;
          const agentList = agentsObj?.list;
          if (Array.isArray(agentList)) {
            const agentCfg = agentList.find(a => a.id === agentId);
            if (agentCfg && agentCfg.workspace) {
              workspaceDir = agentCfg.workspace
            } else {
              api.logger.warn(`Agent workspace not found for: ${agentId}`);
              res.statusCode = 404;
              res.setHeader("Content-Type", "text/plain");
              res.end("Agent workspace not found");
              return true;
            }
          }
        }

        const sessionId = urlObj.searchParams.get("session");
        let normalizedSessionId = sessionId ? sessionId.replace(/[^a-zA-Z0-9-_]/g, "_") : null;
        if (sessionId?.includes(":")) {
          const keyParts = sessionId.split(":");
          const sessionUuid = keyParts[keyParts.length - 1];
          normalizedSessionId = sessionUuid.replace(/[^a-zA-Z0-9-_]/g, "_");
        }

        const archivesDir = path.join(workspaceDir, "archives");
        const rawPath = urlObj.pathname;
        let subPath = decodeURIComponent(rawPath).replace(/^\/v1\/archives\/?/, "");

        if (!subPath && normalizedSessionId) {
          subPath = normalizedSessionId;
        }

        const segments = subPath.split("/").filter(s => !!s);
        if (segments.length > 0 && segments[0].includes(":")) {
          const keyParts = segments[0].split(":");
          const sessionUuid = keyParts[keyParts.length - 1];
          segments[0] = sessionUuid.replace(/[^a-zA-Z0-9-_]/g, "_");
          subPath = segments.join("/");
        }

        const targetPath = path.resolve(archivesDir, subPath);

        const relative = path.relative(archivesDir, targetPath);
        if (relative.startsWith("..") || path.isAbsolute(relative)) {
          api.logger.warn(`Path traversal attempt blocked: ${targetPath}`);
          res.statusCode = 403;
          res.end("Forbidden");
          return true;
        }

        let stat: fs.Stats;
        try {
          stat = await fs.promises.stat(targetPath);
        } catch (e: any) {
          if (e.code === "ENOENT") {
            res.statusCode = 404;
            res.end("Not Found");
            return true;
          }
          throw e;
        }

        if (stat.isDirectory()) {
          const entries = await fs.promises.readdir(targetPath, { withFileTypes: true });

          const files = entries.map(entry => ({
            name: entry.name,
            type: entry.isDirectory() ? "directory" : entry.isFile() ? "file" : "other"
          }));

          files.sort((a, b) => b.name.localeCompare(a.name));

          let displaySubPath = "/";
          if (normalizedSessionId) {
            const sessionDir = path.join(archivesDir, normalizedSessionId);
            displaySubPath = path.relative(sessionDir, targetPath);
          } else {
            const pathParts = subPath.split("/").filter(p => !!p);
            if (pathParts.length > 0) {
              const sessionDir = path.join(archivesDir, pathParts[0]);
              displaySubPath = path.relative(sessionDir, targetPath);
            }
          }
          if (displaySubPath === "." || !displaySubPath) displaySubPath = "/";

          res.statusCode = 200;
          res.setHeader("Content-Type", "application/json");
          res.end(
            JSON.stringify({
              path: displaySubPath,
              contents: files
            })
          );
          return true;
        }

        const ext = path.extname(targetPath).toLowerCase();
        const mimeType = ARCHIVES_MIME_MAP[ext] || "application/octet-stream";
        res.setHeader("Content-Type", mimeType);

        const stream = fs.createReadStream(targetPath);
        res.statusCode = 200;
        stream.pipe(res);
        return true;
      } catch (err: any) {
        api.logger.error(`Error serving archive file: ${err.message}`);
        res.statusCode = 500;
        res.end("Internal Server Error");
        return true;
      }
    }
  });

  api.on("after_tool_call", async (event, ctx) => {
    if (event.error) return;

    const toolName = event.toolName.toLowerCase();
    const isFileModification =
      toolName.includes("write") ||
      toolName.includes("edit") ||
      toolName.includes("replace");

    if (!isFileModification) return;

    const filePathInfo = event.params?.path || event.params?.file || event.params?.filename;
    if (!filePathInfo || typeof filePathInfo !== "string") return;

    try {
      const workspaceDir = api.resolvePath(".") || process.cwd();
      const archivesBaseDir = path.join(workspaceDir, "archives");
      const sourcePath = path.resolve(workspaceDir, filePathInfo);

      const relToWorkspace = path.relative(workspaceDir, sourcePath);
      if (relToWorkspace.startsWith("..") || path.isAbsolute(relToWorkspace)) {
        return;
      }

      let stat;
      try {
        stat = await fs.promises.stat(sourcePath);
      } catch {
        return;
      }

      if (!stat.isFile()) return;

      const sessionKey = ctx.sessionKey || "";
      const sessionParts = sessionKey.split(":");
      const sessionUuid = sessionParts[sessionParts.length - 1] || ctx.sessionId || "unknown-uuid";
      const sessionIdSafe = sessionUuid.replace(/[^a-zA-Z0-9-_]/g, "_");

      const originalFileName = path.basename(sourcePath);
      const isPlan = originalFileName.toLowerCase() === "plan.md";
      const sanitizedFileName = sanitizeFileName(originalFileName);

      const pathSegments = relToWorkspace.split(path.sep);
      const isInArchives = pathSegments[0] === "archives";

      const timestampRegex = /^\d{4}-\d{2}-\d{2}-\d{2}-\d{2}-\d{2}$/;

      let isPathCompliant = false;

      if (isPlan) {
        isPathCompliant =
          isInArchives &&
          pathSegments.length === 3 &&
          pathSegments[1] === sessionIdSafe &&
          pathSegments[2] === "PLAN.md";
      } else {
        isPathCompliant =
          isInArchives &&
          pathSegments.length === 4 &&
          pathSegments[1] === sessionIdSafe &&
          timestampRegex.test(pathSegments[2]) &&
          pathSegments[3] === sanitizedFileName;
      }

      if (isPathCompliant) {
        return;
      }

      const timestamp = formatTimestamp(new Date());
      let targetArchiveDir: string;
      let finalFileName: string = sanitizedFileName;

      if (isPlan) {
        targetArchiveDir = path.join(archivesBaseDir, sessionIdSafe);
        finalFileName = "PLAN.md";
      } else {
        targetArchiveDir = path.join(archivesBaseDir, sessionIdSafe, timestamp);
      }

      await fs.promises.mkdir(targetArchiveDir, { recursive: true });

      const finalTargetInWorkspace = path.join(targetArchiveDir, finalFileName);

      if (sourcePath !== finalTargetInWorkspace) {
        try {
          await fs.promises.copyFile(sourcePath, finalTargetInWorkspace);
          api.logger.info(
            `Copied non-compliant file from ${relToWorkspace} to compliant path ${path.relative(workspaceDir, finalTargetInWorkspace)}`
          );
        } catch (err: any) {
          api.logger.warn(`Failed to copy file to compliant workspace path: ${err.message}`);
        }
      }

      api.logger.info(
        `Processed ${isPlan ? "PLAN" : "archived"} file: ${originalFileName} in ${path.relative(archivesBaseDir, targetArchiveDir)}`
      );
    } catch (err: any) {
      api.logger.error(`Failed to handle archive naming: ${err.message}`);
    }
  });
}
