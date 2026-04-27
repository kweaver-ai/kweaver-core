import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { afterEach, beforeEach, describe, expect, it } from "vitest";

import {
  SkillUninstallError,
  skillUninstallErrorHttpStatus,
  uninstallSkillFromRepo
} from "./skills-uninstall.js";

describe("skills-uninstall", () => {
  let repoSkillsDir: string;

  beforeEach(() => {
    repoSkillsDir = fs.mkdtempSync(path.join(os.tmpdir(), "dip-uninstall-repo-"));
  });

  afterEach(() => {
    fs.rmSync(repoSkillsDir, { recursive: true, force: true });
  });

  it("maps error codes to HTTP status", () => {
    expect(skillUninstallErrorHttpStatus("NOT_FOUND")).toBe(404);
    expect(skillUninstallErrorHttpStatus("INVALID_NAME")).toBe(400);
  });

  it("removes a skill directory under repoSkillsDir", () => {
    const dir = path.join(repoSkillsDir, "weather");
    fs.mkdirSync(dir, { recursive: true });
    fs.writeFileSync(path.join(dir, "SKILL.md"), "# x");

    const result = uninstallSkillFromRepo("weather", repoSkillsDir);

    expect(result).toEqual({ name: "weather" });
    expect(fs.existsSync(dir)).toBe(false);
  });



  it("rejects when skill is missing", () => {
    expect(() =>
      uninstallSkillFromRepo("nope", repoSkillsDir)
    ).toThrow(SkillUninstallError);
    try {
      uninstallSkillFromRepo("nope", repoSkillsDir);
    } catch (e: unknown) {
      expect(e).toMatchObject({ code: "NOT_FOUND" });
    }
  });

  it("rejects invalid names", () => {
    expect(() =>
      uninstallSkillFromRepo("../x", repoSkillsDir)
    ).toThrow(SkillUninstallError);
  });
});
