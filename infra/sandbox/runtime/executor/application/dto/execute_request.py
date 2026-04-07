"""
Execute Request DTO

Data transfer object for execution requests from HTTP layer.
"""

from pydantic import BaseModel, Field


class ExecuteRequestDTO(BaseModel):
    """
    Request DTO for code execution.

    Maps HTTP request to domain ExecutionRequest value object.
    """

    execution_id: str = Field(..., description="Unique execution identifier")
    session_id: str = Field(..., description="Session identifier")
    code: str = Field(..., description="AWS Lambda handler function code")
    language: str = Field(..., description="Programming language")
    event: dict = Field(default_factory=dict, description="Business data passed to handler")
    timeout: int = Field(default=300, description="Timeout in seconds", ge=1, le=3600)
    env_vars: dict = Field(default_factory=dict, description="Environment variables")

    def to_domain(self) -> "executor.domain.value_objects.ExecutionRequest":
        """
        Convert DTO to domain ExecutionRequest value object.

        Returns:
            ExecutionRequest value object
        """
        from executor.domain.value_objects import ExecutionRequest

        return ExecutionRequest(
            code=self.code,
            language=self.language,
            timeout=self.timeout,
            execution_id=self.execution_id,
            session_id=self.session_id,
            event=self.event,
            env_vars=self.env_vars,
        )


class ExecuteResponseDTO(BaseModel):
    """
    Response DTO for code execution.

    Maps domain ExecutionResult to HTTP response.
    """

    execution_id: str
    status: str
    message: str
    error: str | None = None

    @classmethod
    def from_domain(
        cls,
        execution_id: str,
        result: "executor.domain.value_objects.ExecutionResult",
    ) -> "ExecuteResponseDTO":
        """
        Create DTO from domain ExecutionResult.

        Args:
            execution_id: Unique execution identifier
            result: ExecutionResult value object

        Returns:
            ExecuteResponseDTO
        """
        return cls(
            execution_id=execution_id,
            status=result.status.value,
            message="Execution completed",
            error=result.error,
        )


class HealthResponseDTO(BaseModel):
    """
    Health check response DTO.
    """

    status: str = "healthy"
    version: str = "1.0.0"
    uptime_seconds: float | None = None
    active_executions: int | None = None
