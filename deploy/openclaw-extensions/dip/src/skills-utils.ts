import path from "node:path";
import fs from "node:fs";
import os from "node:os";
import type { OpenClawConfig, OpenClawPluginApi } from "openclaw/plugin-sdk";

export type SkillStatusEntry = {
  name: string;
  description: string;
  source: string;
  bundled: boolean;
  filePath: string;
  baseDir: string;
  skillKey: string;
  disabled: boolean;
  blockedByAllowlist: boolean;
  eligible: boolean;
  emoji?: string;
  primaryEnv?: string;
  homepage?: string;
  always?: boolean;
};

export function resolveDefaultAgentId(config: OpenClawConfig): string {
  const agents = (config.agents as any)?.list || [];
  const defaultAgent = agents.find((a: any) => a.default);
  return defaultAgent?.id || "default";
}

export function resolveAgentWorkspaceDir(
  config: OpenClawConfig,
  agentId: string,
  api: OpenClawPluginApi
): string {
  const agents = (config.agents as any)?.list || [];
  const agent = agents.find((a: any) => a.id === agentId);
  const workspace = agent?.workspace;
  if (workspace) {
    return api.resolvePath(workspace);
  }
  const stateDir = api.runtime.state.resolveStateDir();
  return path.join(stateDir, "workspaces", agentId);
}

/**
 * Simplified skill discovery for the dip extension.
 * Scans both the agent's workspace and the global managed skills directory.
 */
export function buildWorkspaceSkillStatus(
  workspaceDir: string,
  opts?: { config?: OpenClawConfig; api?: OpenClawPluginApi }
): { skills: SkillStatusEntry[] } {
  const skills: SkillStatusEntry[] = [];
  const discoveryPaths = new Set<string>();

  // 1. Agent's workspace skills
  discoveryPaths.add(path.join(workspaceDir, "skills"));

  // 2. Global managed skills
  const stateDir = opts?.api?.runtime.state.resolveStateDir() || 
                   path.join(os.homedir(), ".openclaw");
  discoveryPaths.add(path.join(stateDir, "skills"));

  // 3. Personal agents skills
  discoveryPaths.add(path.join(os.homedir(), ".agents", "skills"));

  // 4. Project agents skills (workspace/.agents/skills)
  discoveryPaths.add(path.join(workspaceDir, ".agents", "skills"));

  for (const skillsDir of discoveryPaths) {
    if (fs.existsSync(skillsDir) && fs.statSync(skillsDir).isDirectory()) {
      const entries = fs.readdirSync(skillsDir, { withFileTypes: true });
      for (const entry of entries) {
        const fullPath = path.join(skillsDir, entry.name);
        let isDir = entry.isDirectory();
        
        // Handle symlinks pointing to directories
        if (entry.isSymbolicLink()) {
          try {
            isDir = fs.statSync(fullPath).isDirectory();
          } catch {
            isDir = false;
          }
        }

        if (isDir) {
          const skillMd = path.join(fullPath, "SKILL.md");
          if (fs.existsSync(skillMd)) {
            const name = entry.name;
            const skillKey = name;
            
            // Avoid duplicates; higher priority sources (like workspace) already added
            if (skills.some(s => s.name === name)) continue;

            const skillConfig = (opts?.config?.skills as any)?.entries?.[name];
            const disabled = skillConfig?.enabled === false;
            
            let source = "managed";
            if (skillsDir.includes("workspaces") || skillsDir.includes("agents")) {
              source = "workspace";
            } else if (skillsDir.includes("bundled") || skillsDir.endsWith("skills")) {
              // Heuristic: if it's the root skills dir or env-override, it's bundled
              if (!skillsDir.includes(".openclaw") && !skillsDir.includes(".agents")) {
                source = "bundled";
              }
            }

            skills.push({
              name,
              description: "", 
              source,
              bundled: source === "bundled",
              filePath: skillMd,
              baseDir: fullPath,
              skillKey,
              disabled,
              blockedByAllowlist: false,
              eligible: !disabled,
              always: false,
            });
          }
        }
      }
    }
  }

  return { skills };
}
