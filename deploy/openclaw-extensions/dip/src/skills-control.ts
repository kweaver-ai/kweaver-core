import fs from "node:fs";
import path from "node:path";
import type { IncomingMessage, ServerResponse } from "node:http";
import { type OpenClawPluginApi } from "openclaw/plugin-sdk";
import { resolveAgentWorkspaceDir } from "./skills-utils.js";
import { ARCHIVES_MIME_MAP } from "./archives-utils.js";
import { discoverSkillNames, discoverSkillStatus } from "./skills-discovery.js";
import {
  installSkillFromZipBuffer,
  SkillInstallError,
  skillInstallErrorHttpStatus
} from "./skills-install.js";
import {
  SkillUninstallError,
  skillUninstallErrorHttpStatus,
  uninstallSkillFromRepo
} from "./skills-uninstall.js";
import {
  type SkillTreeEntry,
  SkillTreeError,
  listSkillTreeEntries,
  readSkillFilePreview,
  resolveSkillFilePath
} from "./skills-tree.js";

/** Maximum `.skill` upload size for the Gateway install route (bytes). */
const MAX_SKILL_INSTALL_BYTES = 32 * 1024 * 1024;

/**
 * Reads the raw request body with a hard size cap.
 *
 * @param req Incoming HTTP request.
 * @param maxBytes Maximum allowed body length.
 * @returns Concatenated body buffer.
 */
function readRequestBodyLimited(
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
        reject(
          new SkillInstallError(
            "TOO_LARGE",
            `Request body exceeds ${maxBytes} bytes`
          )
        );
        return;
      }
      chunks.push(chunk);
    });
    req.on("end", () => {
      resolve(Buffer.concat(chunks));
    });
    req.on("error", reject);
  });
}

/**
 * Registers skills CLI command, `/v1/config/agents/skills/install`, `/v1/config/agents/skills/{name}`,
 * and `/v1/config/agents/skills` HTTP routes.
 *
 * @param api OpenClaw plugin API.
 * @param repoRoot Repository / studio root (parent of `skills/`).
 * @param bundledSkillsDir Plugin-relative bundled skills directory.
 */
export function registerSkillsControl(
  api: OpenClawPluginApi
): void {

  api.registerHttpRoute({
    path: "/v1/config/agents/skills/install",
    match: "prefix",
    auth: "gateway",
    handler: async (req: IncomingMessage, res: ServerResponse) => {
      if (req.method !== "POST") {
        res.statusCode = 405;
        res.setHeader("Content-Type", "application/json");
        res.end(
          JSON.stringify({
            error: "Method not allowed. Use POST with a raw application/zip body."
          })
        );
        return true;
      }

      const url = new URL(req.url || "", "http://localhost");
      const overwrite = url.searchParams.get("overwrite") === "true";
      const nameParam = url.searchParams.get("name");
      const agentId = url.searchParams.get("agent");
      const name =
        nameParam !== null && nameParam.trim().length > 0
          ? nameParam.trim()
          : undefined;

      try {
        const config = await api.runtime.config.loadConfig();
        const effectiveAgentId = agentId && agentId.trim().length > 0 ? agentId.trim() : null;
        const repoSkillsDir = effectiveAgentId
          ? path.join(resolveAgentWorkspaceDir(config, effectiveAgentId, api), "skills")
          : path.join(api.runtime.state.resolveStateDir(), "skills");

        const body = await readRequestBodyLimited(req, MAX_SKILL_INSTALL_BYTES);
        const result = installSkillFromZipBuffer(body, repoSkillsDir, {
          overwrite,
          maxBytes: MAX_SKILL_INSTALL_BYTES,
          name
        });
        res.statusCode = 200;
        res.setHeader("Content-Type", "application/json");
        res.end(
          JSON.stringify({
            name: result.name,
            skillPath: result.skillPath
          })
        );
        return true;
      } catch (e: unknown) {
        if (e instanceof SkillInstallError) {
          res.statusCode = skillInstallErrorHttpStatus(e.code);
          res.setHeader("Content-Type", "application/json");
          res.end(JSON.stringify({ error: e.message, code: e.code }));
          return true;
        }
        api.logger.error?.(`dip skills install failed: ${String(e)}`);
        res.statusCode = 500;
        res.setHeader("Content-Type", "application/json");
        res.end(
          JSON.stringify({
            error: e instanceof Error ? e.message : String(e)
          })
        );
        return true;
      }
    }
  });

  api.registerCommand({
    name: "skills-manage",
    description: "Manage agent skills (list, enable, disable)",
    acceptsArgs: true,
    handler: async (ctx: any): Promise<any> => {
      const args = (ctx.args || "").trim().split(/\s+/);
      const sub = args[0]?.toLowerCase();

      if (sub === "list") {
        const skills = discoverSkillStatus(ctx.config, api);
        if (skills.length === 0) return { text: "No skills discovered." };

        const rows = skills.map(s => {
          let status = "✓ ready";
          if (s.disabled) status = "❌ disabled";
          else if (s.blockedByAllowlist) status = "🚫 blocked";
          else if (!s.eligible) status = "✗ missing";

          const icon = s.emoji ? `${s.emoji} ` : "";
          const source = s.source;
          const desc = s.description ? s.description.replace(/(\n|\r)+/g, " ").slice(0, 100) : "";

          return `| ${status} | ${icon}${s.name} | ${source} | ${desc} |`;
        });

        const header = "| Status | Skill | Source | Description |\n|---|---|---|---|";
        return { text: "Available skills:\n\n" + header + "\n" + rows.join("\n") };
      }

      if (sub === "enable" || sub === "disable") {
        const skillName = args[1];
        if (!skillName) return { text: `Usage: /skills-manage ${sub} <name>` };

        const currentConfig = await api.runtime.config.loadConfig();
        const nextCfg = JSON.parse(JSON.stringify(currentConfig));

        if (!nextCfg.skills) nextCfg.skills = {};
        if (!nextCfg.skills.entries) nextCfg.skills.entries = {};
        if (!nextCfg.skills.entries[skillName]) nextCfg.skills.entries[skillName] = {};

        const enabled = sub === "enable";
        nextCfg.skills.entries[skillName].enabled = enabled;

        await api.runtime.config.writeConfigFile(nextCfg);
        return { text: `Global skill "${skillName}" is now ${enabled ? "enabled" : "disabled"}.` };
      }

      return { text: "Usage: /skills-manage [list | enable <name> | disable <name>]" };
    }
  });

  api.registerHttpRoute({
    path: "/v1/config/agents/skills",
    match: "prefix",
    auth: "gateway",
    handler: async (req: IncomingMessage, res: ServerResponse) => {
      const url = new URL(req.url || "", "http://localhost");
      const routeTarget = extractSkillRouteTarget(url.pathname);
      const isBasePath = routeTarget === undefined;

      if (req.method === "DELETE") {
        const name = routeTarget?.name;

        if (name === undefined || routeTarget?.action !== "detail") {
          res.statusCode = 400;
          res.setHeader("Content-Type", "application/json");
          res.end(
            JSON.stringify({
              error: "Path parameter name is required"
            })
          );
          return true;
        }

        try {
          const agentId = url.searchParams.get("agent");
          const config = await api.runtime.config.loadConfig();
          const effectiveAgentId = agentId && agentId.trim().length > 0 ? agentId.trim() : null;
          const repoSkillsDir = effectiveAgentId
            ? path.join(resolveAgentWorkspaceDir(config, effectiveAgentId, api), "skills")
            : path.join(api.runtime.state.resolveStateDir(), "skills");

          const result = uninstallSkillFromRepo(
            name,
            repoSkillsDir
          );
          res.statusCode = 200;
          res.setHeader("Content-Type", "application/json");
          res.end(JSON.stringify({ name: result.name }));
          return true;
        } catch (e: unknown) {
          if (e instanceof SkillUninstallError) {
            res.statusCode = skillUninstallErrorHttpStatus(e.code);
            res.setHeader("Content-Type", "application/json");
            res.end(JSON.stringify({ error: e.message, code: e.code }));
            return true;
          }
          api.logger.error?.(`dip skills uninstall failed: ${String(e)}`);
          res.statusCode = 500;
          res.setHeader("Content-Type", "application/json");
          res.end(
            JSON.stringify({
              error: e instanceof Error ? e.message : String(e)
            })
          );
          return true;
        }
      }

      if (req.method === "GET" && routeTarget?.action === "tree") {
        try {
          const config = await api.runtime.config.loadConfig();
          const entries = resolveSkillTreeEntries(config, api, routeTarget.name);
          res.statusCode = 200;
          res.setHeader("Content-Type", "application/json");
          res.end(JSON.stringify({ name: routeTarget.name, entries }));
          return true;
        } catch (e: unknown) {
          if (e instanceof SkillTreeError) {
            res.statusCode = e.code === "SKILL_NOT_FOUND" ? 404 : 400;
            res.setHeader("Content-Type", "application/json");
            res.end(JSON.stringify({ error: e.message, code: e.code }));
            return true;
          }
          api.logger.error?.(`dip skills tree failed: ${String(e)}`);
          res.statusCode = 500;
          res.setHeader("Content-Type", "application/json");
          res.end(
            JSON.stringify({
              error: e instanceof Error ? e.message : String(e)
            })
          );
          return true;
        }
      }

      if (req.method === "GET" && routeTarget?.action === "content") {
        try {
          const config = await api.runtime.config.loadConfig();
          const skillDir = resolveSkillDirectory(config, api, routeTarget.name);
          const previewPath = url.searchParams.get("path") ?? "";
          const result = readSkillFilePreview(skillDir, previewPath);
          res.statusCode = 200;
          res.setHeader("Content-Type", "application/json");
          res.end(JSON.stringify({ name: routeTarget.name, ...result }));
          return true;
        } catch (e: unknown) {
          if (e instanceof SkillTreeError) {
            res.statusCode = mapSkillTreeErrorStatus(e.code);
            res.setHeader("Content-Type", "application/json");
            res.end(JSON.stringify({ error: e.message, code: e.code }));
            return true;
          }
          api.logger.error?.(`dip skills content failed: ${String(e)}`);
          res.statusCode = 500;
          res.setHeader("Content-Type", "application/json");
          res.end(
            JSON.stringify({
              error: e instanceof Error ? e.message : String(e)
            })
          );
          return true;
        }
      }

      if (req.method === "GET" && routeTarget?.action === "download") {
        try {
          const config = await api.runtime.config.loadConfig();
          const skillDir = resolveSkillDirectory(config, api, routeTarget.name);
          const filePath = url.searchParams.get("path") ?? "";
          const file = resolveSkillFilePath(skillDir, filePath);
          const ext = path.extname(file.absolutePath).toLowerCase();
          const mimeType = ARCHIVES_MIME_MAP[ext] || "application/octet-stream";

          res.statusCode = 200;
          res.setHeader("Content-Type", mimeType);
          res.setHeader(
            "Content-Disposition",
            `attachment; filename="${encodeDownloadFileName(path.basename(file.absolutePath))}"`
          );

          fs.createReadStream(file.absolutePath).pipe(res);
          return true;
        } catch (e: unknown) {
          if (e instanceof SkillTreeError) {
            res.statusCode = mapSkillTreeErrorStatus(e.code);
            res.setHeader("Content-Type", "application/json");
            res.end(JSON.stringify({ error: e.message, code: e.code }));
            return true;
          }
          api.logger.error?.(`dip skills download failed: ${String(e)}`);
          res.statusCode = 500;
          res.setHeader("Content-Type", "application/json");
          res.end(
            JSON.stringify({
              error: e instanceof Error ? e.message : String(e)
            })
          );
          return true;
        }
      }

      if (!isBasePath) {
        res.statusCode = 404;
        res.setHeader("Content-Type", "application/json");
        res.end(JSON.stringify({ error: "Not found" }));
        return true;
      }

      if (req.method === "GET") {
        const agentId = url.searchParams.get("agentId");
        const config = await api.runtime.config.loadConfig();

        if (!agentId) {
          res.statusCode = 200;
          res.setHeader("Content-Type", "application/json");
          const responseSkills = discoverSkillNames(config, api);
          res.end(JSON.stringify({ skills: responseSkills }));
          return true;
        }

        const agent = (config.agents?.list as any[])?.find((a: any) => a.id === agentId);
        if (!agent) {
          res.statusCode = 404;
          res.setHeader("Content-Type", "application/json");
          res.end(JSON.stringify({ error: `Agent "${agentId}" not found` }));
          return true;
        }

        res.statusCode = 200;
        res.setHeader("Content-Type", "application/json");
        const agentSkills = agent.skills;
        let responseSkills = agentSkills;
        if (agentSkills === undefined) {
          responseSkills = discoverSkillNames(config, api, [agentId]);
        }
        res.end(JSON.stringify({ agentId, skills: responseSkills }));
        return true;
      }

      if (req.method === "POST" || req.method === "PUT") {
        try {
          const body = await new Promise<any>((resolve, reject) => {
            let data = "";
            req.on("data", (chunk: any) => (data += chunk));
            req.on("end", () => {
              try {
                resolve(JSON.parse(data));
              } catch {
                reject(new Error("Invalid JSON body"));
              }
            });
            req.on("error", reject);
          });

          const { agentId, skills } = body;
          if (!agentId || !Array.isArray(skills)) {
            res.statusCode = 400;
            res.setHeader("Content-Type", "application/json");
            res.end(JSON.stringify({ error: "agentId (string) and skills (array) are required" }));
            return true;
          }

          const currentConfig = await api.runtime.config.loadConfig();
          const nextCfg = JSON.parse(JSON.stringify(currentConfig));

          if (!nextCfg.agents) nextCfg.agents = {};
          if (!nextCfg.agents.list) nextCfg.agents.list = [];

          const agentIndex = (nextCfg.agents.list as any[]).findIndex((a: any) => a.id === agentId);
          if (agentIndex === -1) {
            res.statusCode = 404;
            res.setHeader("Content-Type", "application/json");
            res.end(JSON.stringify({ error: `Agent "${agentId}" not found in list` }));
            return true;
          }

          nextCfg.agents.list[agentIndex].skills = skills;

          await api.runtime.config.writeConfigFile(nextCfg);

          res.statusCode = 200;
          res.setHeader("Content-Type", "application/json");
          res.end(JSON.stringify({ success: true, agentId, skills }));
          return true;
        } catch (e: any) {
          res.statusCode = 500;
          res.setHeader("Content-Type", "application/json");
          res.end(JSON.stringify({ error: e.message }));
          return true;
        }
      }

      res.statusCode = 405;
      res.setHeader("Content-Type", "application/json");
      res.end(JSON.stringify({ error: "Method not allowed. Use GET to read or POST/PUT to update." }));
      return true;
    }
  });
}

function resolveSkillTreeEntries(config: any, api: OpenClawPluginApi, skillName: string): SkillTreeEntry[] {
  return listSkillTreeEntries(resolveSkillDirectory(config, api, skillName));
}

function readSkillDirectoryFromStatusEntry(entry: Record<string, unknown> | undefined): string | undefined {
  if (entry === undefined) {
    return undefined;
  }

  const baseDir = typeof entry.baseDir === "string" ? entry.baseDir.trim() : "";
  if (baseDir.length > 0) {
    return baseDir;
  }

  const filePath = typeof entry.filePath === "string" ? entry.filePath.trim() : "";
  if (filePath.length > 0) {
    return filePath;
  }

  return undefined;
}

function extractSkillRouteTarget(
  pathname: string | null
): { name: string; action: "detail" | "tree" | "content" | "download" } | undefined {
  if (!pathname) {
    return undefined;
  }

  const normalized = pathname.replace(/\/+$/, "");
  const prefix = "/v1/config/agents/skills";

  if (normalized === prefix) {
    return undefined;
  }

  if (!normalized.startsWith(`${prefix}/`)) {
    return { name: "", action: "detail" };
  }

  const suffix = normalized.slice(prefix.length + 1);
  if (suffix.length === 0) {
    return undefined;
  }

  if (suffix.endsWith("/tree")) {
    const name = suffix.slice(0, -"/tree".length);
    return name.length > 0
      ? { name: decodeURIComponent(name), action: "tree" }
      : { name: "", action: "tree" };
  }

  if (suffix.endsWith("/content")) {
    const name = suffix.slice(0, -"/content".length);
    return name.length > 0
      ? { name: decodeURIComponent(name), action: "content" }
      : { name: "", action: "content" };
  }

  if (suffix.endsWith("/download")) {
    const name = suffix.slice(0, -"/download".length);
    return name.length > 0
      ? { name: decodeURIComponent(name), action: "download" }
      : { name: "", action: "download" };
  }

  return { name: decodeURIComponent(suffix), action: "detail" };
}

function resolveSkillDirectory(config: any, api: OpenClawPluginApi, skillName: string): string {
  const normalizedName = skillName.trim();
  if (normalizedName.length === 0) {
    throw new SkillTreeError("INVALID_NAME", "Path parameter name is required");
  }

  const entry = discoverSkillStatus(config, api).find((candidate) => candidate.name === normalizedName);
  const skillDir = readSkillDirectoryFromStatusEntry(entry);

  if (skillDir === undefined) {
    throw new SkillTreeError("SKILL_NOT_FOUND", `Skill not found: ${normalizedName}`);
  }

  return skillDir;
}

function mapSkillTreeErrorStatus(code: SkillTreeError["code"]): number {
  if (code === "SKILL_NOT_FOUND") {
    return 404;
  }
  if (code === "INVALID_NAME" || code === "INVALID_PATH" || code === "NOT_A_FILE") {
    return 400;
  }
  return 500;
}

function encodeDownloadFileName(fileName: string): string {
  return fileName.replace(/["\\\r\n]/g, "_");
}
