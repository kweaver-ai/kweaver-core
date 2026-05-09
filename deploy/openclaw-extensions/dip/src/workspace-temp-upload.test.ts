import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { afterEach, describe, expect, it } from "vitest";

import {
  buildHashedUploadFilename,
  normalizeSessionSegment,
  parseMultipartUploadFile,
  sanitizeUploadFilename,
  writeWorkspaceTempUpload
} from "./workspace-temp-upload.js";

const createdDirs: string[] = [];

afterEach(() => {
  while (createdDirs.length > 0) {
    const dir = createdDirs.pop();
    if (dir) {
      fs.rmSync(dir, { recursive: true, force: true });
    }
  }
});

describe("workspace-temp-upload", () => {
  it("sanitizes filename and strips path separators", () => {
    expect(sanitizeUploadFilename("../../hello world?.txt")).toBe("hello_world_.txt");
    expect(sanitizeUploadFilename("流程支持并行执行.md")).toBe("流程支持并行执行.md");
    expect(sanitizeUploadFilename("")).toBe("upload.bin");
  });

  it("normalizes session ids from full session key", () => {
    expect(
      normalizeSessionSegment("agent:a:user:b:direct:53615cc3-f321-42eb-8eda-1f1e5c301826")
    ).toBe("53615cc3-f321-42eb-8eda-1f1e5c301826");
    expect(normalizeSessionSegment("")).toBeUndefined();
  });

  it("renames file with hash suffix", () => {
    const name = buildHashedUploadFilename("a.txt", Buffer.from("hello", "utf8"));
    expect(name).toMatch(/^a_[0-9a-f]{12}\.txt$/);
  });

  it("keeps unicode filename before hash suffix", () => {
    const name = buildHashedUploadFilename(
      "流程支持并行执行.md",
      Buffer.from("hello", "utf8")
    );
    expect(name).toMatch(/^流程支持并行执行_[0-9a-f]{12}\.md$/);
  });

  it("falls back to upload basename when filename has only extension", () => {
    const name = buildHashedUploadFilename(".md", Buffer.from("hello", "utf8"));
    expect(name).toMatch(/^upload_[0-9a-f]{12}\.md$/);
  });

  it("parses multipart file payload with original filename", () => {
    const boundary = "----testboundary";
    const multipart = Buffer.from(
      [
        `--${boundary}`,
        'Content-Disposition: form-data; name="file"; filename="流程支持并行执行.md"',
        "Content-Type: text/markdown",
        "",
        "hello",
        `--${boundary}--`,
        ""
      ].join("\r\n"),
      "utf8"
    );

    const parsed = parseMultipartUploadFile(
      multipart,
      `multipart/form-data; boundary=${boundary}`
    );

    expect(parsed.filename).toBe("流程支持并行执行.md");
    expect(parsed.payload.toString("utf8")).toBe("hello");
  });

  it("writes upload into workspace tmp root", async () => {
    const workspaceDir = fs.mkdtempSync(path.join(os.tmpdir(), "dip-tmp-upload-"));
    createdDirs.push(workspaceDir);
    const payload = Buffer.from("hello", "utf8");

    const result = await writeWorkspaceTempUpload(workspaceDir, payload, "note.md");

    expect(result.relativePath.startsWith("tmp/")).toBe(true);
    expect(result.name).toMatch(/^note_[0-9a-f]{12}\.md$/);
    expect(result.bytes).toBe(5);
    expect(fs.readFileSync(result.absolutePath, "utf8")).toBe("hello");
  });

  it("writes upload into session-scoped tmp subdirectory", async () => {
    const workspaceDir = fs.mkdtempSync(path.join(os.tmpdir(), "dip-tmp-upload-session-"));
    createdDirs.push(workspaceDir);

    const result = await writeWorkspaceTempUpload(
      workspaceDir,
      Buffer.from("abc", "utf8"),
      "report.pdf",
      "agent:a:user:b:direct:chat-1"
    );

    expect(result.relativePath.startsWith("tmp/chat-1/")).toBe(true);
    expect(result.name).toMatch(/^report_[0-9a-f]{12}\.pdf$/);
    expect(fs.existsSync(result.absolutePath)).toBe(true);
  });

  it("rejects empty upload body", async () => {
    const workspaceDir = fs.mkdtempSync(path.join(os.tmpdir(), "dip-tmp-upload-empty-"));
    createdDirs.push(workspaceDir);

    await expect(
      writeWorkspaceTempUpload(workspaceDir, Buffer.alloc(0), "x.txt")
    ).rejects.toThrow("Upload body is empty");
  });
});
