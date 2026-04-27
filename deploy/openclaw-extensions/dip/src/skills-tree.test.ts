import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { afterEach, describe, expect, it } from "vitest";

import {
  listSkillTreeEntries,
  readSkillFilePreview,
  resolveSkillFilePath,
  SkillTreeError
} from "./skills-tree.js";

describe("listSkillTreeEntries", () => {
  const tempDirs: string[] = [];

  afterEach(() => {
    for (const dir of tempDirs) {
      fs.rmSync(dir, { recursive: true, force: true });
    }
    tempDirs.length = 0;
  });

  it("returns nested files and directories under a skill", () => {
    const root = fs.mkdtempSync(path.join(os.tmpdir(), "dip-skill-tree-"));
    tempDirs.push(root);

    fs.mkdirSync(path.join(root, "docs"), { recursive: true });
    fs.mkdirSync(path.join(root, "templates", "partials"), { recursive: true });
    fs.writeFileSync(path.join(root, "SKILL.md"), "# skill\n");
    fs.writeFileSync(path.join(root, "docs", "guide.md"), "guide\n");
    fs.writeFileSync(path.join(root, "templates", "base.txt"), "base\n");
    fs.writeFileSync(
      path.join(root, "templates", "partials", "header.txt"),
      "header\n"
    );

    expect(listSkillTreeEntries(root)).toEqual([
      {
        name: "docs",
        path: "docs",
        type: "directory",
        children: [
          {
            name: "guide.md",
            path: "docs/guide.md",
            type: "file"
          }
        ]
      },
      {
        name: "SKILL.md",
        path: "SKILL.md",
        type: "file"
      },
      {
        name: "templates",
        path: "templates",
        type: "directory",
        children: [
          {
            name: "base.txt",
            path: "templates/base.txt",
            type: "file"
          },
          {
            name: "partials",
            path: "templates/partials",
            type: "directory",
            children: [
              {
                name: "header.txt",
                path: "templates/partials/header.txt",
                type: "file"
              }
            ]
          }
        ]
      }
    ]);
  });

  it("throws when the skill directory is missing", () => {
    expect(() => listSkillTreeEntries("/tmp/does-not-exist-skill-tree")).toThrowError(
      SkillTreeError
    );
  });

  it("reads one file preview under a skill directory", () => {
    const root = fs.mkdtempSync(path.join(os.tmpdir(), "dip-skill-preview-"));
    tempDirs.push(root);

    fs.mkdirSync(path.join(root, "docs"), { recursive: true });
    fs.writeFileSync(path.join(root, "docs", "guide.md"), "hello\nworld\n");

    expect(readSkillFilePreview(root, "docs/guide.md", 5)).toEqual({
      path: "docs/guide.md",
      content: "hello",
      bytes: 12,
      truncated: true
    });
  });

  it("defaults to SKILL.md when preview path is empty", () => {
    const root = fs.mkdtempSync(path.join(os.tmpdir(), "dip-skill-preview-"));
    tempDirs.push(root);

    fs.writeFileSync(path.join(root, "SKILL.md"), "# skill\n");

    expect(readSkillFilePreview(root, "")).toEqual({
      path: "SKILL.md",
      content: "# skill\n",
      bytes: 8,
      truncated: false
    });
  });

  it("rejects path traversal and directory previews", () => {
    const root = fs.mkdtempSync(path.join(os.tmpdir(), "dip-skill-preview-"));
    tempDirs.push(root);

    fs.mkdirSync(path.join(root, "docs"), { recursive: true });

    expect(() => readSkillFilePreview(root, "../secret.txt")).toThrowError(
      SkillTreeError
    );
    expect(() => readSkillFilePreview(root, "docs")).toThrowError(SkillTreeError);
  });

  it("resolves file path for download", () => {
    const root = fs.mkdtempSync(path.join(os.tmpdir(), "dip-skill-download-"));
    tempDirs.push(root);

    fs.mkdirSync(path.join(root, "docs"), { recursive: true });
    fs.writeFileSync(path.join(root, "docs", "guide.md"), "hello\n");

    expect(resolveSkillFilePath(root, "docs/guide.md")).toEqual({
      absolutePath: path.join(root, "docs", "guide.md"),
      relativePath: "docs/guide.md",
      bytes: 6
    });
  });
});
