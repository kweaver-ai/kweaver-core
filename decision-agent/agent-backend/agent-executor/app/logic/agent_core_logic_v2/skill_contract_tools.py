"""Built-in skill contract Tool implementations for online mode.

Defines three Tool subclasses that bridge model tool-calls to the
agent-operator-integration skill API:
  - FactorySkillLoadTool       -> builtin_skill_load
  - FactorySkillReadTool       -> builtin_skill_read_file
  - FactorySkillExecuteTool    -> builtin_skill_execute_script

This module belongs to the skill business layer (agent_core_logic_v2), NOT to
the generic tool runtime (tool_v2).  It is referenced only from run_dolphin.py,
which injects these tools into the toolkit when skill_enabled=True.

Service dependency: AgentOperatorIntegrationService
  Skill APIs belong to agent-operator-integration.  The service instance is
  injected via the constructor so tests can substitute it without patching
  module-level globals.
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
)

from app.driven.dip.agent_operator_integration_service import (
    AgentOperatorIntegrationService,
    agent_operator_integration_service,
)
from app.logic.agent_core_logic_v2.factory_skill_runtime import (
    load_skill as _runtime_load_skill,
    read_skill_file as _runtime_read_skill_file,
    execute_skill_script as _runtime_execute_skill_script,
)


class _BuiltinSkillTool(Tool):
    """Base class for built-in skill contract tools.

    ``request_headers`` is captured at construction time (once per request)
    so that concurrent async invocations each carry their own identity headers.
    ``service`` is injected at construction time to allow test substitution.
    """

    inputs_schema: Dict[str, Any] = None

    def __init__(
        self,
        request_headers: Optional[Dict[str, str]] = None,
        service: Optional[AgentOperatorIntegrationService] = None,
    ):
        super().__init__(
            name=self._tool_name(),
            description=self._tool_description(),
            inputs=self._tool_inputs(),
            outputs=self._tool_outputs(),
            props={"intervention": False},
        )
        self.inputs_schema = self._tool_inputs_schema()
        self._request_headers: Optional[Dict[str, str]] = request_headers
        self._service: AgentOperatorIntegrationService = (
            service if service is not None else agent_operator_integration_service
        )

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


class FactorySkillLoadTool(_BuiltinSkillTool):
    """builtin_skill_load — Phase 1, always required.

    Calls GET /skills/{skill_id}/content via agent-operator-integration,
    downloads SKILL.md, and returns skill_md_content + file lists.
    """

    def _tool_name(self) -> str:
        return BUILTIN_SKILL_LOAD

    def _tool_description(self) -> str:
        return SKILL_LOAD_DESCRIPTION

    def _tool_inputs(self) -> Dict[str, Any]:
        return {
            "skill_id": {
                "type": "string",
                "description": SKILL_LOAD_INPUTS_SCHEMA["properties"]["skill_id"]["description"],
                "required": True,
            }
        }

    def _tool_outputs(self) -> Dict[str, Any]:
        return SKILL_LOAD_OUTPUTS

    def _tool_inputs_schema(self) -> Dict[str, Any]:
        return SKILL_LOAD_INPUTS_SCHEMA

    async def arun_stream(self, **kwargs) -> AsyncGenerator[Dict[str, Any], None]:
        skill_id: str = kwargs.get("skill_id") or kwargs.get("tool_input", {}).get("skill_id", "")
        result = await _runtime_load_skill(
            self._service, skill_id, request_headers=self._request_headers
        )
        yield result


class FactorySkillReadTool(_BuiltinSkillTool):
    """builtin_skill_read_file — Phase 2, optional.

    Calls POST /skills/{skill_id}/files/read via agent-operator-integration
    and returns the file's full text content.
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
            self._service, skill_id, file_path, request_headers=self._request_headers
        )
        yield result


class FactorySkillExecuteTool(_BuiltinSkillTool):
    """builtin_skill_execute_script — Phase 3, optional.

    Calls POST /skills/{skill_id}/execute via agent-operator-integration and
    returns stdout, stderr, exit_code, duration_ms, artifacts.
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
            self._service, skill_id, entry_shell, request_headers=self._request_headers
        )
        yield result


def build_builtin_skill_tools(
    request_headers: Optional[Dict[str, str]] = None,
    service: Optional[AgentOperatorIntegrationService] = None,
) -> Dict[str, Tool]:
    """Create the three built-in skill Tools and return a name->Tool dict.

    Called from run_dolphin.py (NOT from build_tools) so that the generic
    tool_v2 layer stays free of skill-specific business logic.

    Args:
        request_headers: Per-request HTTP headers forwarded to every service
            call so that concurrent requests each carry their own identity.
        service: AgentOperatorIntegrationService instance; defaults to the
            module-level singleton (can be overridden in tests).
    """
    kwargs = dict(request_headers=request_headers, service=service)
    return {
        BUILTIN_SKILL_LOAD: FactorySkillLoadTool(**kwargs),
        BUILTIN_SKILL_READ_FILE: FactorySkillReadTool(**kwargs),
        BUILTIN_SKILL_EXECUTE_SCRIPT: FactorySkillExecuteTool(**kwargs),
    }
