import { describe, expect, it } from "vitest";
import { formatTimestamp, sanitizeFileName } from "./archives-utils.js";

describe("archives-utils", () => {
  it("formats timestamps with archive directory precision", () => {
    const date = new Date(2026, 2, 25, 3, 4, 5, 678);
    expect(formatTimestamp(date)).toBe("2026-03-25-03-04-05");
  });

  it("sanitizes filenames and preserves lowercase extension", () => {
    expect(sanitizeFileName("My Report #%?.MD")).toBe("my_report.md");
  });

  it("falls back to unnamed when basename is stripped out", () => {
    expect(sanitizeFileName("###.TXT")).toBe("unnamed.txt");
  });
});
