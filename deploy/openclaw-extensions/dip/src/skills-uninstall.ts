import fs from "node:fs";
import path from "node:path";

/**
 * Error codes returned by {@link uninstallSkillFromRepo} for HTTP mapping.
 */
export type SkillUninstallErrorCode = "INVALID_NAME" | "NOT_FOUND";

/**
 * Thrown when a skill cannot be removed from the repository `skills/` tree.
 */
export class SkillUninstallError extends Error {
  /**
   * Creates a structured uninstall error.
   *
   * @param code Machine-readable error code.
   * @param message Human-readable detail.
   */
  public constructor(
    public readonly code: SkillUninstallErrorCode,
    message: string
  ) {
    super(message);
    this.name = "SkillUninstallError";
  }
}

/** Same rule as `skills-install` directory names. */
const SKILL_NAME_RE = /^[a-zA-Z0-9][a-zA-Z0-9._-]*$/;

/**
 * Ensures {@link resolved} is contained under {@link repoSkillsDir}.
 *
 * @param resolved Absolute or normalized path to candidate.
 * @param repoSkillsDir Repository `skills` root.
 * @returns Whether the path stays inside the repo skills directory.
 */
function isPathUnderRepoSkills(
  resolved: string,
  repoSkillsDir: string
): boolean {
  const root = path.resolve(repoSkillsDir);
  const target = path.resolve(resolved);
  return target === root || target.startsWith(root + path.sep);
}

/**
 * Removes `repoSkillsDir/<name>/` (or a `*.skill` entry) if present.
 *
 * @param name Skill id (slug).
 * @param repoSkillsDir Absolute path to `{workspaceDir}/skills`.
 * @returns Confirmation payload.
 */
export function uninstallSkillFromRepo(
  name: string,
  repoSkillsDir: string
): { name: string } {
  const trimmed = name.trim();
  if (trimmed.length === 0 || !SKILL_NAME_RE.test(trimmed)) {
    throw new SkillUninstallError(
      "INVALID_NAME",
      `Invalid skill name "${name}"`
    );
  }

  const repoTarget = path.join(repoSkillsDir, trimmed);
  if (!isPathUnderRepoSkills(repoTarget, repoSkillsDir)) {
    throw new SkillUninstallError("INVALID_NAME", "Invalid skill path");
  }

  if (fs.existsSync(repoTarget)) {
    fs.rmSync(repoTarget, { recursive: true, force: true });
    return { name: trimmed };
  }

  throw new SkillUninstallError(
    "NOT_FOUND",
    `Skill "${trimmed}" is not installed under skills/`
  );
}

/**
 * Maps a {@link SkillUninstallError} code to an HTTP status code.
 *
 * @param code Error code from {@link SkillUninstallError}.
 * @returns Suggested HTTP status.
 */
export function skillUninstallErrorHttpStatus(
  code: SkillUninstallErrorCode
): number {
  switch (code) {
    case "NOT_FOUND":
      return 404;
    default:
      return 400;
  }
}
