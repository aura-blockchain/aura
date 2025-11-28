"""Module for data registry operations."""

from typing import List, Optional, Dict, Any
from datetime import datetime
from ..types import (
    DataItemParams,
    DataItem,
    DataQuery,
    DataUpdateParams,
    DataAccessLog,
    DataStorageInfo,
    DataItemType,
    AccessLevel,
    TxResult,
    GasOptions
)


class DataRegistryModule:
    """Data registry module for decentralized data storage."""

    def __init__(self, client):
        """Initialize data registry module."""
        self.client = client

    async def register_data(
        self,
        params: DataItemParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Register a new data item.

        Args:
            params: Data item parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.owner:
            raise ValueError("Owner address is required")
        if not params.name:
            raise ValueError("Name is required")
        if not params.content:
            raise ValueError("Content is required")

        message = {
            "@type": "/aura.dataregistry.v1beta1.MsgRegisterData",
            "owner": params.owner,
            "name": params.name,
            "data_type": params.data_type.value if isinstance(params.data_type, DataItemType) else params.data_type,
            "content": params.content,
            "access_level": params.access_level.value if isinstance(params.access_level, AccessLevel) else params.access_level,
            "metadata": params.metadata or {},
            "encryption_key": params.encryption_key or "",
            "expiry": params.expiry.isoformat() if params.expiry else None
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def update_data(
        self,
        params: DataUpdateParams,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Update an existing data item.

        Args:
            params: Update parameters
            options: Transaction options

        Returns:
            Transaction result
        """
        if not params.item_id:
            raise ValueError("Item ID is required")

        message = {
            "@type": "/aura.dataregistry.v1beta1.MsgUpdateData",
            "item_id": params.item_id,
            "content": params.content,
            "metadata": params.metadata,
            "access_level": params.access_level.value if params.access_level and isinstance(params.access_level, AccessLevel) else params.access_level,
            "tags": params.tags or []
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def get_data(self, item_id: str) -> Optional[DataItem]:
        """Get a data item by ID.

        Args:
            item_id: Data item ID

        Returns:
            Data item or None
        """
        if not item_id:
            raise ValueError("Item ID is required")

        try:
            data = await self.client.get(f"/aura/dataregistry/v1beta1/data/{item_id}")
            item_data = data.get("item")

            if not item_data:
                return None

            return DataItem(
                id=item_data.get("id", item_id),
                owner=item_data.get("owner", ""),
                name=item_data.get("name", ""),
                data_type=DataItemType(item_data.get("data_type", "text")),
                content_hash=item_data.get("content_hash", ""),
                access_level=AccessLevel(item_data.get("access_level", "public")),
                created_at=datetime.fromisoformat(item_data.get("created_at")) if item_data.get("created_at") else datetime.now(),
                updated_at=datetime.fromisoformat(item_data.get("updated_at")) if item_data.get("updated_at") else datetime.now(),
                size=item_data.get("size", 0),
                version=item_data.get("version", 1),
                metadata=item_data.get("metadata"),
                tags=item_data.get("tags")
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get data: {e}")

    async def delete_data(
        self,
        item_id: str,
        owner: str,
        options: Optional[GasOptions] = None
    ) -> TxResult:
        """Delete a data item.

        Args:
            item_id: Data item ID
            owner: Owner address
            options: Transaction options

        Returns:
            Transaction result
        """
        if not item_id:
            raise ValueError("Item ID is required")
        if not owner:
            raise ValueError("Owner address is required")

        message = {
            "@type": "/aura.dataregistry.v1beta1.MsgDeleteData",
            "item_id": item_id,
            "owner": owner
        }

        return await self.client.tx_builder.sign_and_broadcast([message], options)

    async def list_data(
        self,
        query: Optional[DataQuery] = None,
        limit: int = 100
    ) -> List[DataItem]:
        """List data items with optional filtering.

        Args:
            query: Query parameters
            limit: Maximum number of results

        Returns:
            List of data items
        """
        try:
            params = {"limit": limit}

            if query:
                if query.owner:
                    params["owner"] = query.owner
                if query.data_type:
                    params["data_type"] = query.data_type.value if isinstance(query.data_type, DataItemType) else query.data_type
                if query.tags:
                    params["tags"] = ",".join(query.tags)
                if query.name_pattern:
                    params["name_pattern"] = query.name_pattern
                if query.created_after:
                    params["created_after"] = query.created_after.isoformat()
                if query.created_before:
                    params["created_before"] = query.created_before.isoformat()

            query_str = "&".join([f"{k}={v}" for k, v in params.items()])
            data = await self.client.get(f"/aura/dataregistry/v1beta1/data?{query_str}")

            items = []
            for item_data in data.get("items", []):
                items.append(DataItem(
                    id=item_data.get("id", ""),
                    owner=item_data.get("owner", ""),
                    name=item_data.get("name", ""),
                    data_type=DataItemType(item_data.get("data_type", "text")),
                    content_hash=item_data.get("content_hash", ""),
                    access_level=AccessLevel(item_data.get("access_level", "public")),
                    created_at=datetime.fromisoformat(item_data.get("created_at")) if item_data.get("created_at") else datetime.now(),
                    updated_at=datetime.fromisoformat(item_data.get("updated_at")) if item_data.get("updated_at") else datetime.now(),
                    size=item_data.get("size", 0),
                    version=item_data.get("version", 1),
                    metadata=item_data.get("metadata"),
                    tags=item_data.get("tags")
                ))

            return items
        except Exception as e:
            raise RuntimeError(f"Failed to list data: {e}")

    async def search_data(
        self,
        search_query: str,
        filters: Optional[Dict[str, Any]] = None,
        limit: int = 100
    ) -> List[DataItem]:
        """Search data items.

        Args:
            search_query: Search query string
            filters: Optional filters
            limit: Maximum number of results

        Returns:
            List of matching data items
        """
        if not search_query:
            raise ValueError("Search query is required")

        try:
            params = {
                "q": search_query,
                "limit": limit
            }

            if filters:
                params.update(filters)

            query_str = "&".join([f"{k}={v}" for k, v in params.items()])
            data = await self.client.get(f"/aura/dataregistry/v1beta1/search?{query_str}")

            items = []
            for item_data in data.get("results", []):
                items.append(DataItem(
                    id=item_data.get("id", ""),
                    owner=item_data.get("owner", ""),
                    name=item_data.get("name", ""),
                    data_type=DataItemType(item_data.get("data_type", "text")),
                    content_hash=item_data.get("content_hash", ""),
                    access_level=AccessLevel(item_data.get("access_level", "public")),
                    created_at=datetime.fromisoformat(item_data.get("created_at")) if item_data.get("created_at") else datetime.now(),
                    updated_at=datetime.fromisoformat(item_data.get("updated_at")) if item_data.get("updated_at") else datetime.now(),
                    size=item_data.get("size", 0),
                    version=item_data.get("version", 1),
                    metadata=item_data.get("metadata"),
                    tags=item_data.get("tags")
                ))

            return items
        except Exception as e:
            raise RuntimeError(f"Failed to search data: {e}")

    async def get_data_stats(self, owner: Optional[str] = None) -> DataStorageInfo:
        """Get data storage statistics.

        Args:
            owner: Optional owner filter

        Returns:
            Storage information
        """
        try:
            path = "/aura/dataregistry/v1beta1/stats"
            if owner:
                path += f"?owner={owner}"

            data = await self.client.get(path)
            stats = data.get("stats", {})

            return DataStorageInfo(
                owner=owner or "global",
                total_items=stats.get("total_items", 0),
                total_size=stats.get("total_size", 0),
                used_quota=stats.get("used_quota", 0.0),
                max_quota=stats.get("max_quota", 0),
                items_by_type=stats.get("items_by_type", {})
            )
        except Exception as e:
            raise RuntimeError(f"Failed to get data stats: {e}")

    async def get_access_logs(
        self,
        item_id: str,
        limit: int = 100
    ) -> List[DataAccessLog]:
        """Get access logs for a data item.

        Args:
            item_id: Data item ID
            limit: Maximum number of results

        Returns:
            List of access logs
        """
        if not item_id:
            raise ValueError("Item ID is required")

        try:
            data = await self.client.get(f"/aura/dataregistry/v1beta1/logs/{item_id}?limit={limit}")

            logs = []
            for log_data in data.get("logs", []):
                logs.append(DataAccessLog(
                    item_id=log_data.get("item_id", item_id),
                    accessor=log_data.get("accessor", ""),
                    action=log_data.get("action", ""),
                    timestamp=datetime.fromisoformat(log_data.get("timestamp")) if log_data.get("timestamp") else datetime.now(),
                    success=log_data.get("success", False),
                    ip_address=log_data.get("ip_address")
                ))

            return logs
        except Exception as e:
            raise RuntimeError(f"Failed to get access logs: {e}")
