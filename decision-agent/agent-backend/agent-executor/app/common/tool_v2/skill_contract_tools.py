"""Built-in skill contract Tool implementations for online mode.

Defines three Tool subclasses that wrap the execution-factory skill API:
- FactorySkillLoadTool       -> builtin_skill_load
- FactorySkillReadTool       -> builtin_skill_read_file
- FactorySkillExecuteTool    -> builtin_skill_execute_script

These Tools are injected into the tool_dict by build_tools() and are always
present regardless of what skills the agent has configured.  Their model-visible
names are fixed reserved names: any user-defined tool with the same name is
overwritten by these built-in tools.

Schemas are imported from dolphin.sdk.skill.skill_contracts so that the
executor side never duplicates the contract definition.
"""

from typing import Any, AsyncGenerator, Dict, Optional

from dolphin.core.utils.tools import Tool
from dolphin.sdk.skill.skill_contracts import (
    BUILTIN_SKILL_LOAD,
    BUILTIN_SKILL_READ_FILE,
    BUILTIN_SKILL_EXECUTE_SCRIPT,
    SKILL_LOAD_DESCRIPTION,
    SKILL_READ_FILE_DESCRIPTION,
    SKILL_EXECUTE_SCRIPT_DESCRIPTION,
    SKILL_LOAD_INPUTS_SCHEMA,
    SKILL_READ_FILE_INPUTS_SCHEMA,
    SKILL_EXECUTE_SCRIPT_INPUTS_SCHEMA,
    SKILL_LOAD_OUTPUTS,
    SKILL_READ_FILE_OUTPUTS,
    SKILL_EXECUTE_SCRIPT_OUTPUTS,
    SKILL_LOAD_OPENAI_SCHEMA,
    SKILL_READ_FILE_OPENAI_SCHEMA,
    SKILL_EXECUTE_SCRIPT_OPENAI_SCHEMA,
)

from app.driven.dip.agent_factory_service import agent_factory_service
from app.logic.agent_core_logic_v2.factory_skill_runtime import (
    load_skill as _runtime_load_skill,
    read_skill_file as _runtime_read_skill_file,
    execute_skill_script as _runtime_execute_skill_script,
)


class _BuiltinSkillTool(Tool):
    """Base class for built-in skill contract tools.

    Subclasses must set `name`, `description`, `inputs_schema`, and
    override `arun_stream` to call the appropriate runtime function.

    ``request_headers`` is captured at construction time (once per request,
    inside build_builtin_skill_tools) so that concurrent async invocations
    each carry their own identity headers without mutating the shared
    agent_factory_service singleton.
    """

    # inputs_schema is used by TriditionalToolkit._tool_to_openai_schema
    inputs_schema: Dict[str, Any] = None

    def __init__(self, request_headers: Optional[Dict[str, str]] = None):
        super().__init__(
            name=self._tool_name(),
            description=self._tool_description(),
            inputs=self._tool_inputs(),
            outputs=self._tool_outputs(),
            props={"intervention": False},
        )
        self.inputs_schema = self._tool_inputs_schema()
        # Snapshot of the per-request auth headers; passed explicitly to every
        # service call so the singleton's self.headers is never read at tool
        # invocation time.
        self._request_headers: Optional[Dict[str, str]] = request_headers

    def _tool_name(self) -> str:
        raise NotImplementedError

    def _tool_description(self) -> str:
        raise NotImplementedError

    def _tool_inputs(self) -> Dict[str, Any]:
        raise NotImplementedError

    def _tool_outputs(self) -> Dict[str, Any]:
        raise NotImplementedError

    def _tool_inputs_schema(self) -> Dict[str, Any]:
        raise NotImplementedError

    async def arun_stream(self, **kwargs) -> AsyncGenerator[Dict[str, Any], None]:
        raise NotImplementedError


# ---------------------------------------------------------------------------
# FactorySkillLoadTool
# ---------------------------------------------------------------------------


class FactorySkillLoadTool(_BuiltinSkillTool):
    """Built-in Tool for builtin_skill_load (Phase 1 — always required).

    Calls GET /skills/{skill_id}/content, downloads SKILL.md, and returns
    a unified result containing skill_md_content, available_scripts,
    and available_references.
    """

    def _tool_name(self) -> str:
        return BUILTIN_SKILL_LOAD

    def _tool_description(self) -> str:
        return SKILL_LOAD_DESCRIPTION

    def _tool_inputs(self) -> Dict[str, Any]:
        return {
            "skill_id": {
                "type": "string",
                "description": SKILL_LOAD_INPUTS_SCHEMA["properties"]["skill_id"][
                    "description"
                ],
                "required": True,
            }
        }

    def _tool_outputs(self) -> Dict[str, Any]:
        return SKILL_LOAD_OUTPUTS

    def _tool_inputs_schema(self) -> Dict[str, Any]:
        return SKILL_LOAD_INPUTS_SCHEMA

    async def arun_stream(self, **kwargs) -> AsyncGenerator[Dict[str, Any], None]:
        skill_id: str = kwargs.get("skill_id") or kwargs.get("tool_input", {}).get(
            "skill_id", ""
        )
        result = await _runtime_load_skill(
            agent_factory_service, skill_id, request_headers=self._request_headers
        )
        yield result


# ---------------------------------------------------------------------------
# FactorySkillReadTool
# ---------------------------------------------------------------------------


class FactorySkillReadTool(_BuiltinSkillTool):
    """Built-in Tool for builtin_skill_read_file (Phase 2 — optional).

    Calls POST /skills/{skill_id}/files/read, downloads the target file,
    and returns its full text content.
    """

    def _tool_name(self) -> str:
        return BUILTIN_SKILL_READ_FILE

    def _tool_description(self) -> str:
        return SKILL_READ_FILE_DESCRIPTION

    def _tool_inputs(self) -> Dict[str, Any]:
        props = SKILL_READ_FILE_INPUTS_SCHEMA["properties"]
        return {
            "skill_id": {
                "type": "string",
                "description": props["skill_id"]["description"],
                "required": True,
            },
            "file_path": {
                "type": "string",
                "description": props["file_path"]["description"],
                "required": True,
            },
        }

    def _tool_outputs(self) -> Dict[str, Any]:
        return SKILL_READ_FILE_OUTPUTS

    def _tool_inputs_schema(self) -> Dict[str, Any]:
        return SKILL_READ_FILE_INPUTS_SCHEMA

    async def arun_stream(self, **kwargs) -> AsyncGenerator[Dict[str, Any], None]:
        tool_input = kwargs.get("tool_input", {})
        skill_id: str = kwargs.get("skill_id") or tool_input.get("skill_id", "")
        file_path: str = kwargs.get("file_path") or tool_input.get("file_path", "")
        result = await _runtime_read_skill_file(
            agent_factory_service,
            skill_id,
            file_path,
            request_headers=self._request_headers,
        )
        yield result


# ---------------------------------------------------------------------------
# FactorySkillExecuteTool
# ---------------------------------------------------------------------------


class FactorySkillExecuteTool(_BuiltinSkillTool):
    """Built-in Tool for builtin_skill_execute_script (Phase 3 — optional).

    Calls POST /skills/{skill_id}/execute and returns structured execution
    results (stdout, stderr, exit_code, duration_ms, artifacts).

    The LLM reads the entry_shell command from SKILL.md and passes it
    directly — no path-to-command conversion is performed here.
    """

    def _tool_name(self) -> str:
        return BUILTIN_SKILL_EXECUTE_SCRIPT

    def _tool_description(self) -> str:
        return SKILL_EXECUTE_SCRIPT_DESCRIPTION

    def _tool_inputs(self) -> Dict[str, Any]:
        props = SKILL_EXECUTE_SCRIPT_INPUTS_SCHEMA["properties"]
        return {
            "skill_id": {
                "type": "string",
                "description": props["skill_id"]["description"],
                "required": True,
            },
            "entry_shell": {
                "type": "string",
                "description": props["entry_shell"]["description"],
                "required": True,
            },
        }

    def _tool_outputs(self) -> Dict[str, Any]:
        return SKILL_EXECUTE_SCRIPT_OUTPUTS

    def _tool_inputs_schema(self) -> Dict[str, Any]:
        return SKILL_EXECUTE_SCRIPT_INPUTS_SCHEMA

    async def arun_stream(self, **kwargs) -> AsyncGenerator[Dict[str, Any], None]:
        tool_input = kwargs.get("tool_input", {})
        skill_id: str = kwargs.get("skill_id") or tool_input.get("skill_id", "")
        entry_shell: str = kwargs.get("entry_shell") or tool_input.get("entry_shell", "")
        result = await _runtime_execute_skill_script(
            agent_factory_service,
            skill_id,
            entry_shell,
            request_headers=self._request_headers,
        )
        yield result


# ---------------------------------------------------------------------------
# Factory helper
# ---------------------------------------------------------------------------


def build_builtin_skill_tools(
    request_headers: Optional[Dict[str, str]] = None,
) -> Dict[str, Tool]:
    """Create the three built-in skill Tools and return them as a name->Tool dict.

    tool.py calls this function after building regular user tools and merges
    the result into tool_dict, ensuring these reserved names always win.

    ``request_headers`` must be the per-request auth headers for the current
    call (typically forwarded from run_dolphin's ``headers`` argument).  Each
    Tool instance snapshots these headers at construction time so that
    subsequent ``arun_stream`` invocations pass them directly to every service
    call, avoiding the set_headers() race condition on the shared singleton
    under concurrent async execution.

    Args:
        request_headers: Per-request HTTP headers containing user identity
            fields (x-user-account-id, x-business-domain, etc.)

    Returns:
        Dict mapping each contract name to its Tool instance
    """
    return {
        BUILTIN_SKILL_LOAD: FactorySkillLoadTool(request_headers),
        BUILTIN_SKILL_READ_FILE: FactorySkillReadTool(request_headers),
        BUILTIN_SKILL_EXECUTE_SCRIPT: FactorySkillExecuteTool(request_headers),
    }
