import { spawnSync } from "node:child_process";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { afterEach, beforeEach, describe, expect, it } from "vitest";

import {
  installSkillFromZipBuffer,
  SkillInstallError,
  skillInstallErrorHttpStatus
} from "./skills-install";

/**
 * Builds a minimal zip using the host `zip` or `tar` CLI (no JS zip libraries).
 *
 * @param skillId Skill directory name inside the archive.
 * @param skillMd SKILL.md body.
 * @returns Zip bytes.
 */
function buildSkillZip(skillId: string, skillMd = "# Test skill\n"): Buffer {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), "dip-skills-zip-"));
  try {
    const skillDir = path.join(root, skillId);
    fs.mkdirSync(skillDir, { recursive: true });
    fs.writeFileSync(path.join(skillDir, "SKILL.md"), skillMd);
    const outZip = path.join(root, "out.zip");

    const zipCmd = spawnSync("zip", ["-q", "-r", outZip, skillId], { cwd: root });
    if (zipCmd.status === 0) {
      return fs.readFileSync(outZip);
    }

    const tarCmd = spawnSync("tar", ["-a", "-cf", outZip, skillId], { cwd: root });
    if (tarCmd.status === 0) {
      return fs.readFileSync(outZip);
    }

    throw new Error(
      "Cannot build test zip: install `zip` or a `tar` that supports `-a` zip output"
    );
  } finally {
    fs.rmSync(root, { recursive: true, force: true });
  }
}

/**
 * @param entries Relative directory names under `root`, each gets SKILL.md.
 */
/**
 * Zip with `SKILL.md` at archive root (optional extra files).
 */
function buildFlatZip(skillMd = "# Flat skill\n"): Buffer {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), "dip-skills-flat-"));
  try {
    fs.writeFileSync(path.join(root, "SKILL.md"), skillMd);
    fs.writeFileSync(path.join(root, "extra.txt"), "x");
    const outZip = path.join(root, "out.zip");
    const zipCmd = spawnSync("zip", ["-q", outZip, "SKILL.md", "extra.txt"], { cwd: root });
    if (zipCmd.status === 0) {
      return fs.readFileSync(outZip);
    }
    const tarCmd = spawnSync("tar", ["-a", "-cf", outZip, "SKILL.md", "extra.txt"], {
      cwd: root
    });
    if (tarCmd.status === 0) {
      return fs.readFileSync(outZip);
    }
    throw new Error("Cannot build flat test zip");
  } finally {
    fs.rmSync(root, { recursive: true, force: true });
  }
}

function buildMultiRootZip(entries: string[]): Buffer {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), "dip-skills-zip-"));
  try {
    for (const name of entries) {
      const dir = path.join(root, name);
      fs.mkdirSync(dir, { recursive: true });
      fs.writeFileSync(path.join(dir, "SKILL.md"), "x");
    }
    const outZip = path.join(root, "out.zip");
    const zipCmd = spawnSync("zip", ["-q", "-r", outZip, ...entries], { cwd: root });
    if (zipCmd.status === 0) {
      return fs.readFileSync(outZip);
    }
    const tarCmd = spawnSync("tar", ["-a", "-cf", outZip, ...entries], { cwd: root });
    if (tarCmd.status === 0) {
      return fs.readFileSync(outZip);
    }
    throw new Error("Cannot build multi-root test zip");
  } finally {
    fs.rmSync(root, { recursive: true, force: true });
  }
}

const canPackZip = ((): boolean => {
  try {
    buildSkillZip("probe", "# p\n");
    return true;
  } catch {
    return false;
  }
})();

describe("skills-install", () => {
  let repoSkillsDir: string;

  beforeEach(() => {
    repoSkillsDir = fs.mkdtempSync(path.join(os.tmpdir(), "dip-repo-skills-"));
  });

  afterEach(() => {
    fs.rmSync(repoSkillsDir, { recursive: true, force: true });
  });

  it("maps error codes to HTTP status", () => {
    expect(skillInstallErrorHttpStatus("TOO_LARGE")).toBe(413);
    expect(skillInstallErrorHttpStatus("CONFLICT")).toBe(409);
    expect(skillInstallErrorHttpStatus("BAD_LAYOUT")).toBe(400);
  });

  it("rejects when buffer is empty", () => {
    expect(() => installSkillFromZipBuffer(Buffer.alloc(0), repoSkillsDir)).toThrow(
      SkillInstallError
    );
    try {
      installSkillFromZipBuffer(Buffer.alloc(0), repoSkillsDir);
    } catch (e: unknown) {
      expect(e).toMatchObject({ code: "INVALID_ZIP" });
    }
  });

  it.skipIf(!canPackZip)("rejects when buffer exceeds maxBytes", () => {
    const zip = buildSkillZip("big", "# b\n");
    expect(() =>
      installSkillFromZipBuffer(zip, repoSkillsDir, { maxBytes: 10 })
    ).toThrow(SkillInstallError);
    try {
      installSkillFromZipBuffer(zip, repoSkillsDir, { maxBytes: 10 });
    } catch (e: unknown) {
      expect(e).toMatchObject({ code: "TOO_LARGE" });
    }
  });

  it.skipIf(!canPackZip)("installs a skill directory under repoSkillsDir", () => {
    const zip = buildSkillZip("weather");

    const result = installSkillFromZipBuffer(zip, repoSkillsDir);

    expect(result.name).toBe("weather");
    expect(result.displayName).toBeUndefined();
    expect(result.skillPath).toBe(path.join(repoSkillsDir, "weather"));
    expect(
      fs.readFileSync(path.join(repoSkillsDir, "weather", "SKILL.md"), "utf8")
    ).toContain("Test skill");
  });

  it.skipIf(!canPackZip)("includes displayName from SKILL.md front matter", () => {
    const zip = buildSkillZip(
      "weather",
      ["---", "name: weather", "---", "# body"].join("\n")
    );

    const result = installSkillFromZipBuffer(zip, repoSkillsDir);

    expect(result.name).toBe("weather");
    expect(result.displayName).toBe("weather");
  });

  it.skipIf(!canPackZip)("rejects when skill exists and overwrite is false", () => {
    const zip = buildSkillZip("dup");
    installSkillFromZipBuffer(zip, repoSkillsDir);

    expect(() => installSkillFromZipBuffer(zip, repoSkillsDir)).toThrow(SkillInstallError);
    try {
      installSkillFromZipBuffer(zip, repoSkillsDir);
    } catch (e: unknown) {
      expect(e).toMatchObject({ code: "CONFLICT" });
    }
  });

  it.skipIf(!canPackZip)("replaces an existing skill when overwrite is true", () => {
    const zip1 = buildSkillZip("dup", "v1");
    const zip2 = buildSkillZip("dup", "v2");
    installSkillFromZipBuffer(zip1, repoSkillsDir);

    installSkillFromZipBuffer(zip2, repoSkillsDir, { overwrite: true });

    expect(fs.readFileSync(path.join(repoSkillsDir, "dup", "SKILL.md"), "utf8")).toContain("v2");
  });

  it.skipIf(!canPackZip)("rejects multiple top-level directories", () => {
    const buf = buildMultiRootZip(["a", "b"]);

    expect(() => installSkillFromZipBuffer(buf, repoSkillsDir)).toThrow(SkillInstallError);
    try {
      installSkillFromZipBuffer(buf, repoSkillsDir);
    } catch (e: unknown) {
      expect(e).toMatchObject({ code: "BAD_LAYOUT" });
    }
  });

  it.skipIf(!canPackZip)("installs flat zip when SKILL.md is at archive root and name is set", () => {
    const zip = buildFlatZip(["---", "name: flat-skill", "---", "# hello\n"].join("\n"));

    const result = installSkillFromZipBuffer(zip, repoSkillsDir, { name: "flat-skill" });

    expect(result.name).toBe("flat-skill");
    expect(result.displayName).toBe("flat-skill");
    expect(
      fs.readFileSync(path.join(repoSkillsDir, "flat-skill", "SKILL.md"), "utf8")
    ).toContain("hello");
    expect(fs.readFileSync(path.join(repoSkillsDir, "flat-skill", "extra.txt"), "utf8")).toBe("x");
  });

  it.skipIf(!canPackZip)("rejects when SKILL.md name mismatches directory name", () => {
    const zip = buildSkillZip(
      "weather",
      ["---", "name: other", "---", "# body"].join("\n")
    );

    expect(() => installSkillFromZipBuffer(zip, repoSkillsDir)).toThrow(SkillInstallError);
    try {
      installSkillFromZipBuffer(zip, repoSkillsDir);
    } catch (e: unknown) {
      expect(e).toMatchObject({ code: "INVALID_NAME" });
    }
  });

  it.skipIf(!canPackZip)("rejects flat layout when SKILL.md name mismatches provided slug", () => {
    const zip = buildFlatZip(["---", "name: other", "---", "# body"].join("\n"));

    expect(() =>
      installSkillFromZipBuffer(zip, repoSkillsDir, { name: "flat-skill" })
    ).toThrow(SkillInstallError);
    try {
      installSkillFromZipBuffer(zip, repoSkillsDir, { name: "flat-skill" });
    } catch (e: unknown) {
      expect(e).toMatchObject({ code: "INVALID_NAME" });
    }
  });

  it.skipIf(!canPackZip)("rejects flat zip without name", () => {
    const zip = buildFlatZip();

    expect(() => installSkillFromZipBuffer(zip, repoSkillsDir)).toThrow(SkillInstallError);
    try {
      installSkillFromZipBuffer(zip, repoSkillsDir);
    } catch (e: unknown) {
      expect(e).toMatchObject({ code: "INVALID_NAME" });
    }
  });

  it.skipIf(!canPackZip)("rejects when SKILL.md is missing", () => {
    const root = fs.mkdtempSync(path.join(os.tmpdir(), "dip-skills-zip-"));
    try {
      fs.mkdirSync(path.join(root, "nope"), { recursive: true });
      fs.writeFileSync(path.join(root, "nope", "README.md"), "x");
      const outZip = path.join(root, "out.zip");
      const zipCmd = spawnSync("zip", ["-q", "-r", outZip, "nope"], { cwd: root });
      if (zipCmd.status !== 0) {
        const tarCmd = spawnSync("tar", ["-a", "-cf", outZip, "nope"], { cwd: root });
        if (tarCmd.status !== 0) {
          return;
        }
      }
      const buf = fs.readFileSync(outZip);
      expect(() => installSkillFromZipBuffer(buf, repoSkillsDir)).toThrow(SkillInstallError);
      try {
        installSkillFromZipBuffer(buf, repoSkillsDir);
      } catch (e: unknown) {
        expect(e).toMatchObject({ code: "MISSING_SKILL_MD" });
      }
    } finally {
      fs.rmSync(root, { recursive: true, force: true });
    }
  });
});
