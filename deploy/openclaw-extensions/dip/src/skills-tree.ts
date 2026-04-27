import fs from "node:fs";
import path from "node:path";

export type SkillTreeEntryType = "file" | "directory";

export interface SkillTreeEntry {
  name: string;
  path: string;
  type: SkillTreeEntryType;
  children?: SkillTreeEntry[];
}

export interface SkillFilePreview {
  path: string;
  content: string;
  bytes: number;
  truncated: boolean;
}

export class SkillTreeError extends Error {
  public constructor(
    public readonly code:
      | "INVALID_NAME"
      | "SKILL_NOT_FOUND"
      | "INVALID_PATH"
      | "NOT_A_FILE",
    message: string
  ) {
    super(message);
    this.name = "SkillTreeError";
  }
}

const DEFAULT_SKILL_PREVIEW_MAX_BYTES = 1024 * 1024;

export function listSkillTreeEntries(skillDir: string): SkillTreeEntry[] {
  const normalizedSkillDir = path.resolve(skillDir);

  ensureSkillDirectoryExists(normalizedSkillDir);

  return readTreeEntries(normalizedSkillDir, normalizedSkillDir);
}

export function readSkillFilePreview(
  skillDir: string,
  relativePath: string,
  maxBytes = DEFAULT_SKILL_PREVIEW_MAX_BYTES
): SkillFilePreview {
  const { absolutePath, relativePath: normalizedPath, bytes } =
    resolveSkillFilePath(skillDir, relativePath);

  const previewBytes = fs.readFileSync(absolutePath);
  const content = previewBytes.subarray(0, maxBytes).toString("utf8");

  return {
    path: normalizedPath,
    content,
    bytes,
    truncated: bytes > maxBytes
  };
}

export function resolveSkillFilePath(
  skillDir: string,
  relativePath: string
): { absolutePath: string; relativePath: string; bytes: number } {
  const normalizedSkillDir = path.resolve(skillDir);
  ensureSkillDirectoryExists(normalizedSkillDir);

  const normalizedPath = normalizeRelativeSkillPath(relativePath);
  const absolutePath = path.resolve(normalizedSkillDir, normalizedPath);
  const relativeToRoot = path.relative(normalizedSkillDir, absolutePath);

  if (relativeToRoot.startsWith("..") || path.isAbsolute(relativeToRoot)) {
    throw new SkillTreeError("INVALID_PATH", "File path must stay within the skill directory");
  }

  if (!fs.existsSync(absolutePath)) {
    throw new SkillTreeError("SKILL_NOT_FOUND", `Skill file not found: ${normalizedPath}`);
  }

  const stat = fs.statSync(absolutePath);
  if (!stat.isFile()) {
    throw new SkillTreeError("NOT_A_FILE", `Skill path is not a file: ${normalizedPath}`);
  }

  return {
    absolutePath,
    relativePath: toPosixPath(relativeToRoot),
    bytes: stat.size
  };
}

function readTreeEntries(rootDir: string, currentDir: string): SkillTreeEntry[] {
  const entries = fs.readdirSync(currentDir, { withFileTypes: true });

  return entries
    .sort((a, b) => compareDirents(a.name, b.name))
    .map((entry) => {
      const absolutePath = path.join(currentDir, entry.name);
      const relativePath = path.relative(rootDir, absolutePath);

      if (entry.isDirectory()) {
        return {
          name: entry.name,
          path: toPosixPath(relativePath),
          type: "directory",
          children: readTreeEntries(rootDir, absolutePath)
        } satisfies SkillTreeEntry;
      }

      return {
        name: entry.name,
        path: toPosixPath(relativePath),
        type: "file"
      } satisfies SkillTreeEntry;
    });
}

function compareDirents(left: string, right: string): number {
  return left.localeCompare(right, "en");
}

function toPosixPath(value: string): string {
  return value.split(path.sep).join("/");
}

function ensureSkillDirectoryExists(skillDir: string): void {
  if (!fs.existsSync(skillDir) || !fs.statSync(skillDir).isDirectory()) {
    throw new SkillTreeError(
      "SKILL_NOT_FOUND",
      `Skill directory does not exist: ${skillDir}`
    );
  }
}

function normalizeRelativeSkillPath(value: string): string {
  const trimmed = value.trim();
  if (trimmed.length === 0) {
    return "SKILL.md";
  }

  const normalized = trimmed.replace(/\\/g, "/").replace(/^\/+/, "");
  if (normalized.length === 0 || normalized === "." || normalized === "..") {
    throw new SkillTreeError("INVALID_PATH", "Query parameter path is invalid");
  }

  return normalized;
}
