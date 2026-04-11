import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const { buildWorkspaceSkillStatus, resolveAgentWorkspaceDir, resolveDefaultAgentId } = vi.hoisted(() => ({
  buildWorkspaceSkillStatus: vi.fn(),
  resolveAgentWorkspaceDir: vi.fn().mockReturnValue("/mock/agent/workspace"),
  resolveDefaultAgentId: vi.fn().mockReturnValue("default-agent")
}));

vi.mock("./skills-utils.js", () => ({
  buildWorkspaceSkillStatus,
  resolveAgentWorkspaceDir,
  resolveDefaultAgentId
}));

import { discoverSkillNames } from "./skills-discovery.js";

describe("skills-discovery", () => {
  const mockApi = {
    resolvePath: vi.fn((p) => p),
    runtime: {
      state: {
        resolveStateDir: vi.fn().mockReturnValue("/mock/state")
      },
      config: {
        loadConfig: vi.fn()
      }
    }
  } as any;

  beforeEach(() => {
    buildWorkspaceSkillStatus.mockReset();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("returns sorted unique skill names from SDK discovery", () => {
    buildWorkspaceSkillStatus.mockReturnValue({
      skills: [
        { name: "schedule-plan" },
        { name: "archive-protocol" },
        { name: "schedule-plan" }
      ]
    });

    const result = discoverSkillNames({ any: "config" } as any, mockApi);

    expect(result).toEqual(["archive-protocol", "schedule-plan"]);
    expect(buildWorkspaceSkillStatus).toHaveBeenCalledWith(
      expect.any(String),
      { config: { any: "config" }, api: mockApi }
    );
  });

  it("passes agentIds to SDK discovery", () => {
    buildWorkspaceSkillStatus.mockReturnValue({
      skills: [
        { name: "contextloader" },
        { name: "schedule-plan" },
        { name: "schedule-plan" }
      ]
    });

    const result = discoverSkillNames({ cfg: true } as any, mockApi, ["agent-1"]);

    expect(result).toEqual(["contextloader", "schedule-plan"]);
  });
});
