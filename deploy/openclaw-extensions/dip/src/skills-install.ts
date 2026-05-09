import { spawnSync } from "node:child_process";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";

/**
 * Error codes returned by {@link installSkillFromZipBuffer} for HTTP mapping.
 */
export type SkillInstallErrorCode =
  | "INVALID_ZIP"
  | "BAD_LAYOUT"
  | "MISSING_SKILL_MD"
  | "INVALID_NAME"
  | "CONFLICT"
  | "TOO_LARGE";

/**
 * Thrown when a `.skill` zip cannot be installed to the repository `skills/` tree.
 */
export class SkillInstallError extends Error {
  /**
   * Creates a structured install error.
   *
   * @param code Machine-readable error code.
   * @param message Human-readable detail.
   */
  public constructor(
    public readonly code: SkillInstallErrorCode,
    message: string
  ) {
    super(message);
    this.name = "SkillInstallError";
  }
}

const DEFAULT_MAX_BYTES = 32 * 1024 * 1024;

/** Allowed skill directory names (slug style). */
const SKILL_NAME_RE = /^[a-zA-Z0-9][a-zA-Z0-9._-]*$/;

const IGNORE_TOP_LEVEL = new Set(["__MACOSX", ".DS_Store"]);

/**
 * Validates a single path segment for safe extraction.
 *
 * @param name Directory or file name.
 * @returns Whether the segment is safe.
 */
function isSafeSegment(name: string): boolean {
  return name.length > 0 && name !== "." && name !== ".." && !name.includes("/") && !name.includes("\\");
}

/**
 * Extracts a zip archive into `destDir` using only the system `tar` or `unzip` command (no npm zip libraries).
 *
 * @param zipPath Absolute path to the `.zip` file on disk.
 * @param destDir Directory to extract into (must exist or be creatable).
 */
function extractZipWithSystemTools(zipPath: string, destDir: string): void {
  fs.mkdirSync(destDir, { recursive: true });

  // Use --force-local to prevent tar from treating 'C:' in Windows paths as a remote host.
  const tar = spawnSync("tar", ["--force-local", "-xf", zipPath, "-C", destDir], {
    encoding: "utf8",
    maxBuffer: 32 * 1024 * 1024
  });

  if (tar.status === 0) {
    return;
  }

  const unzip = spawnSync("unzip", ["-q", "-o", zipPath, "-d", destDir], {
    encoding: "utf8",
    maxBuffer: 32 * 1024 * 1024
  });

  // unzip returns 0 for success, 1 for warnings (like 'extra bytes' which are often harmless).
  if (unzip.status === 0 || unzip.status === 1) {
    return;
  }

  const hint =
    tar.error?.message ??
    unzip.error?.message ??
    [tar.stderr, unzip.stderr].filter(Boolean).join(" ").trim();

  throw new SkillInstallError(
    "INVALID_ZIP",
    `Could not extract zip (need 'tar' or 'unzip' on PATH). ${hint}`
  );
}

/**
 * Installs a `.skill` zip (OpenClaw skill bundle) into `repoSkillsDir/<name>/`.
 *
 * Uses the host `tar` or `unzip` binary to read the archive — no third-party zip npm packages.
 *
 * Layout **either**:
 *
 * - **Nested**: exactly one top-level directory whose name is the skill id, containing `SKILL.md`; or
 * - **Flat**: `SKILL.md` at the archive root (any number of sibling files/folders). Requires
 *   `options.name` (slug). The entire extracted tree is copied to `skills/<name>/`.
 *
 * @param zipBuffer Raw zip bytes.
 * @param repoSkillsDir Absolute path to the repository `skills` directory (`path.join(repoRoot, "skills")`).
 * @param options Optional size limit, overwrite, and `name` (required for flat layout).
 * @returns The installed skill id and absolute destination path.
 */
export function installSkillFromZipBuffer(
  zipBuffer: Buffer,
  repoSkillsDir: string,
  options?: { overwrite?: boolean; maxBytes?: number; name?: string }
): { name: string; skillPath: string; displayName?: string } {
  const maxBytes = options?.maxBytes ?? DEFAULT_MAX_BYTES;
  if (zipBuffer.length > maxBytes) {
    throw new SkillInstallError(
      "TOO_LARGE",
      `Zip exceeds maximum size of ${maxBytes} bytes`
    );
  }

  if (zipBuffer.length === 0) {
    throw new SkillInstallError("INVALID_ZIP", "Empty upload");
  }

  const workDir = fs.mkdtempSync(path.join(os.tmpdir(), "dip-skill-install-"));
  const zipPath = path.join(workDir, "upload.zip");
  const extractRoot = path.join(workDir, "extracted");

  try {
    fs.mkdirSync(extractRoot, { recursive: true });
    fs.writeFileSync(zipPath, zipBuffer);

    try {
      extractZipWithSystemTools(zipPath, extractRoot);
    } catch (e: unknown) {
      if (e instanceof SkillInstallError) {
        throw e;
      }
      throw new SkillInstallError(
        "INVALID_ZIP",
        e instanceof Error ? e.message : String(e)
      );
    }

    const topLevel = fs.readdirSync(extractRoot, { withFileTypes: true }).filter(
      (dirent) => !IGNORE_TOP_LEVEL.has(dirent.name)
    );

    if (topLevel.length === 0) {
      throw new SkillInstallError("BAD_LAYOUT", "Zip has no usable top-level entries");
    }

    const rootSkillMdPath = path.join(extractRoot, "SKILL.md");
    const hasRootSkillMd = fs.existsSync(rootSkillMdPath);

    if (hasRootSkillMd) {
      const installName = options?.name?.trim();
      if (installName === undefined || installName.length === 0) {
        throw new SkillInstallError(
          "INVALID_NAME",
          "Archive has SKILL.md at zip root; pass `name` (e.g. ?name=my-skill) to choose the skills/ folder name"
        );
      }
      if (!isSafeSegment(installName) || !SKILL_NAME_RE.test(installName)) {
        throw new SkillInstallError(
          "INVALID_NAME",
          `Invalid skill name "${installName}" (expected a slug such as "my-skill")`
        );
      }

      const displayName = readSkillNameFromFrontMatter(rootSkillMdPath);
      if (displayName !== undefined && displayName.trim() !== installName) {
        throw new SkillInstallError(
          "INVALID_NAME",
          `SKILL.md name "${displayName}" must match slug "${installName}"`
        );
      }

      assertExtractedPathsContained(extractRoot, extractRoot);

      const dest = path.join(repoSkillsDir, installName);
      if (fs.existsSync(dest)) {
        if (!options?.overwrite) {
          throw new SkillInstallError(
            "CONFLICT",
            `Skill "${installName}" already exists under skills/`
          );
        }
        fs.rmSync(dest, { recursive: true, force: true });
      }

      fs.mkdirSync(repoSkillsDir, { recursive: true });
      fs.cpSync(extractRoot, dest, { recursive: true });

      return { name: installName, skillPath: dest, displayName };
    }

    if (topLevel.length > 1) {
      throw new SkillInstallError(
        "BAD_LAYOUT",
        "Zip has multiple top-level entries but no SKILL.md at archive root; use a single top-level folder with SKILL.md inside, or add SKILL.md at the zip root and pass name"
      );
    }

    const only = topLevel[0];
    if (!only.isDirectory()) {
      throw new SkillInstallError(
        "BAD_LAYOUT",
        "Zip must contain SKILL.md at archive root (flat layout + name) or one top-level directory with SKILL.md inside"
      );
    }

    const skillName = only.name;
    if (!isSafeSegment(skillName) || !SKILL_NAME_RE.test(skillName)) {
      throw new SkillInstallError(
        "INVALID_NAME",
        `Invalid skill name "${skillName}" (expected a slug such as "my-skill")`
      );
    }

    const extractedRoot = path.join(extractRoot, skillName);
    const skillMdPath = path.join(extractedRoot, "SKILL.md");
    if (!fs.existsSync(skillMdPath)) {
      throw new SkillInstallError(
        "MISSING_SKILL_MD",
        `Missing ${skillName}/SKILL.md in archive`
      );
    }

    const displayName = readSkillNameFromFrontMatter(skillMdPath);
    if (displayName !== undefined && displayName.trim() !== skillName) {
      throw new SkillInstallError(
        "INVALID_NAME",
        `SKILL.md name "${displayName}" must match slug "${skillName}"`
      );
    }

    assertExtractedPathsContained(extractRoot, extractedRoot);

    const dest = path.join(repoSkillsDir, skillName);
    if (fs.existsSync(dest)) {
      if (!options?.overwrite) {
        throw new SkillInstallError(
          "CONFLICT",
          `Skill "${skillName}" already exists under skills/`
        );
      }
      fs.rmSync(dest, { recursive: true, force: true });
    }

    fs.mkdirSync(repoSkillsDir, { recursive: true });
    fs.cpSync(extractedRoot, dest, { recursive: true });

    return { name: skillName, skillPath: dest, displayName };
  } finally {
    fs.rmSync(workDir, { recursive: true, force: true });
  }
}

/**
 * Ensures every file under the extracted skill tree stays inside the temp sandbox.
 *
 * @param extractRoot Parent directory passed to `tar` / `unzip` (`-C` / `-d`).
 * @param skillDir Path to `<extractRoot>/<name>`.
 */
function assertExtractedPathsContained(extractRoot: string, skillDir: string): void {
  const resolvedRoot = path.resolve(skillDir);
  const resolvedExtract = path.resolve(extractRoot);
  if (!resolvedRoot.startsWith(resolvedExtract + path.sep) && resolvedRoot !== resolvedExtract) {
    throw new SkillInstallError("BAD_LAYOUT", "Extracted paths escape extract directory");
  }

  const walk = (dir: string): void => {
    const entries = fs.readdirSync(dir, { withFileTypes: true });
    for (const entry of entries) {
      const full = path.join(dir, entry.name);
      const resolved = path.resolve(full);
      if (!resolved.startsWith(resolvedRoot + path.sep) && resolved !== resolvedRoot) {
        throw new SkillInstallError("BAD_LAYOUT", "Zip slip or invalid nested paths");
      }
      if (entry.isDirectory()) {
        walk(full);
      }
    }
  };

  walk(skillDir);
}

/**
 * Maps a {@link SkillInstallError} code to an HTTP status code.
 *
 * @param code Error code from {@link SkillInstallError}.
 * @returns Suggested HTTP status.
 */
export function skillInstallErrorHttpStatus(code: SkillInstallErrorCode): number {
  switch (code) {
    case "TOO_LARGE":
      return 413;
    case "CONFLICT":
      return 409;
    default:
      return 400;
  }
}
function readSkillNameFromFrontMatter(skillMdPath: string): string | undefined {
  try {
    const contents = fs.readFileSync(skillMdPath, "utf8");
    const lines = contents.split(/\r?\n/);
    if (lines[0]?.trim() !== "---") {
      return undefined;
    }
    for (let i = 1; i < lines.length; i++) {
      const line = lines[i];
      if (line.trim() === "---") {
        break;
      }
      const match = line.match(/^name:\s*(.+)$/);
      if (match) {
        return match[1].trim().replace(/^["']|["']$/g, "");
      }
    }
  } catch {
    return undefined;
  }
  return undefined;
}
