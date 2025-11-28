"""Module for confidence score operations."""

from typing import List, Optional, Dict, Any
from datetime import datetime
from ..types import (
    ConfidenceScoreParams,
    ConfidenceScore,
    ScoreHistory,
    ScoreBreakdown,
    RewardParams,
    SlashParams,
    IRCompletionParams,
    TxResult,
    GasOptions
)


class ConfidenceScoreModule:
    """Confidence score module for node scoring and reputation."""

    def __init__(self, client):
        """Initialize confidence score module."""
        self.client = client

    async def get_confidence_score(
        self,
        address: str,
        include_history: bool = False,
        include_breakdown: bool = False
    ) -> ConfidenceScore:
        """Get confidence score for an address.

        Args:
            address: Address to query
            include_history: Include historical data
            include_breakdown: Include score breakdown

        Returns:
            Confidence score
        """
        if not address:
            raise ValueError("Address is required")

        try:
            params = {
                "address": address,
                "include_history": include_history,
                "include_breakdown": include_breakdown
            }
            query_str = "&".join([f"{k}={v}".lower() for k, v in params.items()])
            data = await self.client.get(f"/aura/confidencescore/v1beta1/score?{query_str}")

            score_data = data.get("score", {})

            return ConfidenceScore(
                address=score_data.get("address", address),
                score=score_data.get("score", 0.0),
                rank=score_data.get("rank", 0),
                total_participants=score_data.get("total_participants", 0),
                last_updated=datetime.fromisoformat(score_data.get("last_updated")) if score_data.get("last_updated") else datetime.now(),
                factors=score_data.get("factors"),
                metadata=score_data.get("metadata")
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get confidence score: {e}")

    async def update_score(
        self,
        address: str,
        score_delta: float,
        reason: str,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Update confidence score (admin only).

        Args:
            address: Address to update
            score_delta: Score change (positive or negative)
            reason: Reason for update
            options: Transaction options

        Returns:
            Transaction result
        """
        if not address:
            raise ValueError("Address is required")
        if not reason:
            raise ValueError("Reason is required")

        message = {
            "@type": "/aura.confidencescore.v1beta1.MsgUpdateScore",
            "address": address,
            "score_delta": str(score_delta),
            "reason": reason
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_score_history(
        self,
        address: str,
        start_date: Optional[datetime] = None,
        end_date: Optional[datetime] = None
    ) -> ScoreHistory:
        """Get historical confidence scores.

        Args:
            address: Address to query
            start_date: Start date filter
            end_date: End date filter

        Returns:
            Score history
        """
        if not address:
            raise ValueError("Address is required")

        try:
            params = {"address": address}
            if start_date:
                params["start_date"] = start_date.isoformat()
            if end_date:
                params["end_date"] = end_date.isoformat()

            query_str = "&".join([f"{k}={v}" for k, v in params.items()])
            data = await self.client.get(f"/aura/confidencescore/v1beta1/history?{query_str}")

            history_data = data.get("history", {})

            return ScoreHistory(
                address=address,
                scores=history_data.get("scores", []),
                start_date=datetime.fromisoformat(history_data.get("start_date")) if history_data.get("start_date") else (start_date or datetime.now()),
                end_date=datetime.fromisoformat(history_data.get("end_date")) if history_data.get("end_date") else (end_date or datetime.now()),
                average_score=history_data.get("average_score", 0.0),
                peak_score=history_data.get("peak_score", 0.0),
                lowest_score=history_data.get("lowest_score", 0.0)
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get score history: {e}")

    async def reward_node(
        self,
        params: RewardParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Reward a node for completing inclusion routine.

        Args:
            params: Reward parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.address:
            raise ValueError("Address is required")
        if not params.routine_id:
            raise ValueError("Routine ID is required")

        message = {
            "@type": "/aura.confidencescore.v1beta1.MsgRewardNode",
            "address": params.address,
            "routine_id": params.routine_id,
            "completion_rate": str(params.completion_rate)
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def slash_node(
        self,
        params: SlashParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Slash a node for misbehavior.

        Args:
            params: Slashing parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.address:
            raise ValueError("Address is required")
        if not params.reason:
            raise ValueError("Reason is required")
        if not params.penalty_amount:
            raise ValueError("Penalty amount is required")

        message = {
            "@type": "/aura.confidencescore.v1beta1.MsgSlashNode",
            "address": params.address,
            "reason": params.reason,
            "penalty_amount": params.penalty_amount,
            "evidence": params.evidence or ""
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_score_params(self) -> Dict[str, Any]:
        """Get confidence score module parameters.

        Returns:
            Score parameters
        """
        try:
            data = await self.client.get("/aura/confidencescore/v1beta1/params")
            return data.get("params", {})
        except Exception as e:
            raise RuntimeError(f"Failed to get score params: {e}")

    async def get_score_breakdown(self, address: str) -> ScoreBreakdown:
        """Get detailed score breakdown for an address.

        Args:
            address: Address to query

        Returns:
            Score breakdown
        """
        if not address:
            raise ValueError("Address is required")

        try:
            data = await self.client.get(f"/aura/confidencescore/v1beta1/breakdown/{address}")
            breakdown_data = data.get("breakdown", {})

            return ScoreBreakdown(
                participation_score=breakdown_data.get("participation_score", 0.0),
                reliability_score=breakdown_data.get("reliability_score", 0.0),
                reputation_score=breakdown_data.get("reputation_score", 0.0),
                activity_score=breakdown_data.get("activity_score", 0.0),
                penalties=breakdown_data.get("penalties", 0.0),
                bonuses=breakdown_data.get("bonuses", 0.0)
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get score breakdown: {e}")

    async def record_ir_completion(
        self,
        params: IRCompletionParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Record inclusion routine completion.

        Args:
            params: Completion parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.routine_id:
            raise ValueError("Routine ID is required")
        if not params.participant:
            raise ValueError("Participant address is required")

        message = {
            "@type": "/aura.confidencescore.v1beta1.MsgRecordCompletion",
            "routine_id": params.routine_id,
            "participant": params.participant,
            "success": params.success,
            "execution_time": params.execution_time,
            "gas_used": params.gas_used
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_leaderboard(self, limit: int = 100) -> List[ConfidenceScore]:
        """Get confidence score leaderboard.

        Args:
            limit: Maximum number of results

        Returns:
            List of top scores
        """
        try:
            data = await self.client.get(f"/aura/confidencescore/v1beta1/leaderboard?limit={limit}")

            leaderboard = []
            for score_data in data.get("leaderboard", []):
                leaderboard.append(ConfidenceScore(
                    address=score_data.get("address", ""),
                    score=score_data.get("score", 0.0),
                    rank=score_data.get("rank", 0),
                    total_participants=score_data.get("total_participants", 0),
                    last_updated=datetime.fromisoformat(score_data.get("last_updated")) if score_data.get("last_updated") else datetime.now(),
                    factors=score_data.get("factors"),
                    metadata=score_data.get("metadata")
                ))

            return leaderboard
        except Exception as e:
            raise RuntimeError(f"Failed to get leaderboard: {e}")
