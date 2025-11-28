"""Module for inclusion routines operations."""

from typing import List, Optional, Dict, Any
from datetime import datetime
from ..types import (
    InclusionRoutineParams,
    InclusionRoutine,
    RoutineExecution,
    RoutineUpdateParams,
    RoutineRateLimitParams,
    RoutinePrerequisite,
    RoutineRegistry,
    RoutineStatus,
    RoutinePriority,
    TxResult,
    GasOptions
)


class InclusionRoutinesModule:
    """Inclusion routines module for managing on-chain computation routines."""

    def __init__(self, client):
        """Initialize inclusion routines module."""
        self.client = client

    async def create_routine(
        self,
        params: InclusionRoutineParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Create a new inclusion routine.

        Args:
            params: Routine parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.creator:
            raise ValueError("Creator address is required")
        if not params.name:
            raise ValueError("Routine name is required")
        if not params.code:
            raise ValueError("Routine code is required")

        message = {
            "@type": "/aura.inclusionroutines.v1beta1.MsgCreateRoutine",
            "creator": params.creator,
            "name": params.name,
            "description": params.description,
            "code": params.code,
            "priority": params.priority.value if isinstance(params.priority, RoutinePriority) else params.priority,
            "max_gas": params.max_gas,
            "prerequisites": params.prerequisites or [],
            "schedule": params.schedule or "",
            "metadata": params.metadata or {}
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def execute_routine(
        self,
        routine_id: str,
        executor: str,
        params: Optional[Dict[str, Any]] = None,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Execute an inclusion routine.

        Args:
            routine_id: Routine ID
            executor: Executor address
            params: Optional execution parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not routine_id:
            raise ValueError("Routine ID is required")
        if not executor:
            raise ValueError("Executor address is required")

        message = {
            "@type": "/aura.inclusionroutines.v1beta1.MsgExecuteRoutine",
            "routine_id": routine_id,
            "executor": executor,
            "params": params or {}
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_routine(self, routine_id: str) -> Optional[InclusionRoutine]:
        """Get an inclusion routine by ID.

        Args:
            routine_id: Routine ID

        Returns:
            Inclusion routine or None
        """
        if not routine_id:
            raise ValueError("Routine ID is required")

        try:
            data = await self.client.get(f"/aura/inclusionroutines/v1beta1/routines/{routine_id}")
            routine_data = data.get("routine")

            if not routine_data:
                return None

            return InclusionRoutine(
                routine_id=routine_data.get("routine_id", routine_id),
                creator=routine_data.get("creator", ""),
                name=routine_data.get("name", ""),
                description=routine_data.get("description", ""),
                status=RoutineStatus(routine_data.get("status", "active")),
                priority=RoutinePriority(routine_data.get("priority", "medium")),
                code_hash=routine_data.get("code_hash", ""),
                created_at=datetime.fromisoformat(routine_data.get("created_at")) if routine_data.get("created_at") else datetime.now(),
                updated_at=datetime.fromisoformat(routine_data.get("updated_at")) if routine_data.get("updated_at") else datetime.now(),
                executions=routine_data.get("executions", 0),
                success_rate=routine_data.get("success_rate", 0.0),
                average_gas=routine_data.get("average_gas", 0),
                last_execution=datetime.fromisoformat(routine_data.get("last_execution")) if routine_data.get("last_execution") else None
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get routine: {e}")

    async def list_routines(
        self,
        creator: Optional[str] = None,
        status: Optional[RoutineStatus] = None,
        limit: int = 100
    ) -> List[InclusionRoutine]:
        """List inclusion routines.

        Args:
            creator: Optional creator filter
            status: Optional status filter
            limit: Maximum number of results

        Returns:
            List of routines
        """
        try:
            params = {"limit": limit}
            if creator:
                params["creator"] = creator
            if status:
                params["status"] = status.value if isinstance(status, RoutineStatus) else status

            query_str = "&".join([f"{k}={v}" for k, v in params.items()])
            data = await self.client.get(f"/aura/inclusionroutines/v1beta1/routines?{query_str}")

            routines = []
            for routine_data in data.get("routines", []):
                routines.append(InclusionRoutine(
                    routine_id=routine_data.get("routine_id", ""),
                    creator=routine_data.get("creator", ""),
                    name=routine_data.get("name", ""),
                    description=routine_data.get("description", ""),
                    status=RoutineStatus(routine_data.get("status", "active")),
                    priority=RoutinePriority(routine_data.get("priority", "medium")),
                    code_hash=routine_data.get("code_hash", ""),
                    created_at=datetime.fromisoformat(routine_data.get("created_at")) if routine_data.get("created_at") else datetime.now(),
                    updated_at=datetime.fromisoformat(routine_data.get("updated_at")) if routine_data.get("updated_at") else datetime.now(),
                    executions=routine_data.get("executions", 0),
                    success_rate=routine_data.get("success_rate", 0.0),
                    average_gas=routine_data.get("average_gas", 0),
                    last_execution=datetime.fromisoformat(routine_data.get("last_execution")) if routine_data.get("last_execution") else None
                ))

            return routines
        except Exception as e:
            raise RuntimeError(f"Failed to list routines: {e}")

    async def update_routine(
        self,
        params: RoutineUpdateParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Update an inclusion routine.

        Args:
            params: Update parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.routine_id:
            raise ValueError("Routine ID is required")

        message = {
            "@type": "/aura.inclusionroutines.v1beta1.MsgUpdateRoutine",
            "routine_id": params.routine_id,
            "code": params.code,
            "priority": params.priority.value if params.priority and isinstance(params.priority, RoutinePriority) else params.priority,
            "schedule": params.schedule,
            "status": params.status.value if params.status and isinstance(params.status, RoutineStatus) else params.status
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def delete_routine(
        self,
        routine_id: str,
        creator: str,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Delete an inclusion routine.

        Args:
            routine_id: Routine ID
            creator: Creator address
            options: Transaction options

        Returns:
            Transaction result
        """
        if not routine_id:
            raise ValueError("Routine ID is required")
        if not creator:
            raise ValueError("Creator address is required")

        message = {
            "@type": "/aura.inclusionroutines.v1beta1.MsgDeleteRoutine",
            "routine_id": routine_id,
            "creator": creator
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_execution_history(
        self,
        routine_id: str,
        limit: int = 100
    ) -> List[RoutineExecution]:
        """Get execution history for a routine.

        Args:
            routine_id: Routine ID
            limit: Maximum number of results

        Returns:
            List of executions
        """
        if not routine_id:
            raise ValueError("Routine ID is required")

        try:
            data = await self.client.get(f"/aura/inclusionroutines/v1beta1/executions/{routine_id}?limit={limit}")

            executions = []
            for exec_data in data.get("executions", []):
                executions.append(RoutineExecution(
                    execution_id=exec_data.get("execution_id", ""),
                    routine_id=exec_data.get("routine_id", routine_id),
                    executor=exec_data.get("executor", ""),
                    status=exec_data.get("status", ""),
                    started_at=datetime.fromisoformat(exec_data.get("started_at")) if exec_data.get("started_at") else datetime.now(),
                    completed_at=datetime.fromisoformat(exec_data.get("completed_at")) if exec_data.get("completed_at") else None,
                    gas_used=exec_data.get("gas_used", 0),
                    result=exec_data.get("result"),
                    error=exec_data.get("error")
                ))

            return executions
        except Exception as e:
            raise RuntimeError(f"Failed to get execution history: {e}")

    async def set_rate_limit(
        self,
        params: RoutineRateLimitParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Set rate limit for a routine.

        Args:
            params: Rate limit parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.routine_id:
            raise ValueError("Routine ID is required")

        message = {
            "@type": "/aura.inclusionroutines.v1beta1.MsgSetRateLimit",
            "routine_id": params.routine_id,
            "max_executions_per_block": params.max_executions_per_block,
            "cooldown_blocks": params.cooldown_blocks
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_registry_stats(self) -> RoutineRegistry:
        """Get registry statistics.

        Returns:
            Registry statistics
        """
        try:
            data = await self.client.get("/aura/inclusionroutines/v1beta1/stats")
            stats = data.get("stats", {})

            return RoutineRegistry(
                total_routines=stats.get("total_routines", 0),
                active_routines=stats.get("active_routines", 0),
                total_executions=stats.get("total_executions", 0),
                average_success_rate=stats.get("average_success_rate", 0.0),
                top_routines=stats.get("top_routines", [])
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get registry stats: {e}")
