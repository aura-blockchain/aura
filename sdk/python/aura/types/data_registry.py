"""Type definitions for DataRegistry module."""

from dataclasses import dataclass
from typing import Optional, List, Dict, Any
from datetime import datetime
from enum import Enum


class DataItemType(Enum):
    """Data item types."""

    TEXT = "text"
    JSON = "json"
    BINARY = "binary"
    ENCRYPTED = "encrypted"


class AccessLevel(Enum):
    """Access control levels."""

    PUBLIC = "public"
    PRIVATE = "private"
    RESTRICTED = "restricted"


@dataclass
class DataItemParams:
    """Parameters for creating data item."""

    owner: str
    name: str
    data_type: DataItemType
    content: str
    access_level: AccessLevel
    metadata: Optional[Dict[str, Any]] = None
    encryption_key: Optional[str] = None
    expiry: Optional[datetime] = None


@dataclass
class DataItem:
    """Data registry item."""

    id: str
    owner: str
    name: str
    data_type: DataItemType
    content_hash: str
    access_level: AccessLevel
    created_at: datetime
    updated_at: datetime
    size: int
    version: int
    metadata: Optional[Dict[str, Any]] = None
    tags: Optional[List[str]] = None


@dataclass
class DataQuery:
    """Data query parameters."""

    owner: Optional[str] = None
    data_type: Optional[DataItemType] = None
    tags: Optional[List[str]] = None
    name_pattern: Optional[str] = None
    created_after: Optional[datetime] = None
    created_before: Optional[datetime] = None


@dataclass
class DataUpdateParams:
    """Parameters for updating data item."""

    item_id: str
    content: Optional[str] = None
    metadata: Optional[Dict[str, Any]] = None
    access_level: Optional[AccessLevel] = None
    tags: Optional[List[str]] = None


@dataclass
class DataAccessLog:
    """Data access log entry."""

    item_id: str
    accessor: str
    action: str
    timestamp: datetime
    success: bool
    ip_address: Optional[str] = None


@dataclass
class DataStorageInfo:
    """Storage information."""

    owner: str
    total_items: int
    total_size: int
    used_quota: float
    max_quota: int
    items_by_type: Dict[str, int]
