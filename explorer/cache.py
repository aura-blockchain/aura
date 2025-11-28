"""
Caching layer for block explorer
Multi-tier caching with Redis and in-memory support
"""

import json
import logging
import time
from typing import Any, Optional, Callable
from functools import wraps
from dataclasses import dataclass
import hashlib

logger = logging.getLogger(__name__)


@dataclass
class CacheEntry:
    """Cache entry with metadata"""
    key: str
    value: Any
    timestamp: float
    ttl: int
    hit_count: int = 0


class MemoryCache:
    """In-memory LRU cache"""

    def __init__(self, max_size: int = 1000):
        self.max_size = max_size
        self.cache: dict[str, CacheEntry] = {}
        self.access_order: list[str] = []

    def get(self, key: str) -> Optional[Any]:
        """Get value from cache"""
        if key in self.cache:
            entry = self.cache[key]

            # Check if expired
            if time.time() - entry.timestamp > entry.ttl:
                self.delete(key)
                return None

            # Update access order (LRU)
            if key in self.access_order:
                self.access_order.remove(key)
            self.access_order.append(key)

            # Update hit count
            entry.hit_count += 1

            return entry.value

        return None

    def set(self, key: str, value: Any, ttl: int = 300) -> None:
        """Set value in cache"""
        # Evict if at capacity
        if len(self.cache) >= self.max_size and key not in self.cache:
            self._evict_lru()

        entry = CacheEntry(
            key=key, value=value, timestamp=time.time(), ttl=ttl, hit_count=0
        )

        self.cache[key] = entry

        if key in self.access_order:
            self.access_order.remove(key)
        self.access_order.append(key)

    def delete(self, key: str) -> None:
        """Delete key from cache"""
        if key in self.cache:
            del self.cache[key]
        if key in self.access_order:
            self.access_order.remove(key)

    def clear(self) -> None:
        """Clear all cache"""
        self.cache.clear()
        self.access_order.clear()

    def _evict_lru(self) -> None:
        """Evict least recently used item"""
        if self.access_order:
            lru_key = self.access_order.pop(0)
            if lru_key in self.cache:
                del self.cache[lru_key]

    def get_stats(self) -> dict:
        """Get cache statistics"""
        total_hits = sum(entry.hit_count for entry in self.cache.values())
        return {
            "size": len(self.cache),
            "max_size": self.max_size,
            "total_hits": total_hits,
            "keys": list(self.cache.keys()),
        }


class RedisCache:
    """Redis-based cache (stub - would use actual Redis)"""

    def __init__(self, redis_url: str = "redis://localhost:6379"):
        self.redis_url = redis_url
        self.enabled = False
        # In production, you would:
        # import redis
        # self.client = redis.from_url(redis_url)
        # self.enabled = True

    def get(self, key: str) -> Optional[Any]:
        """Get value from Redis"""
        if not self.enabled:
            return None

        try:
            # value = self.client.get(key)
            # if value:
            #     return json.loads(value)
            return None
        except Exception as e:
            logger.error(f"Redis get error: {e}")
            return None

    def set(self, key: str, value: Any, ttl: int = 300) -> None:
        """Set value in Redis"""
        if not self.enabled:
            return

        try:
            # serialized = json.dumps(value)
            # self.client.setex(key, ttl, serialized)
            pass
        except Exception as e:
            logger.error(f"Redis set error: {e}")

    def delete(self, key: str) -> None:
        """Delete key from Redis"""
        if not self.enabled:
            return

        try:
            # self.client.delete(key)
            pass
        except Exception as e:
            logger.error(f"Redis delete error: {e}")

    def clear(self) -> None:
        """Clear all cache"""
        if not self.enabled:
            return

        try:
            # self.client.flushdb()
            pass
        except Exception as e:
            logger.error(f"Redis clear error: {e}")


class MultiTierCache:
    """Multi-tier cache with memory and Redis"""

    def __init__(
        self,
        memory_cache: Optional[MemoryCache] = None,
        redis_cache: Optional[RedisCache] = None,
    ):
        self.l1_cache = memory_cache or MemoryCache(max_size=1000)
        self.l2_cache = redis_cache
        self.stats = {"l1_hits": 0, "l2_hits": 0, "misses": 0}

    def get(self, key: str) -> Optional[Any]:
        """Get value from cache (L1 -> L2)"""
        # Try L1 (memory) first
        value = self.l1_cache.get(key)
        if value is not None:
            self.stats["l1_hits"] += 1
            return value

        # Try L2 (Redis) if available
        if self.l2_cache:
            value = self.l2_cache.get(key)
            if value is not None:
                self.stats["l2_hits"] += 1
                # Promote to L1
                self.l1_cache.set(key, value)
                return value

        self.stats["misses"] += 1
        return None

    def set(self, key: str, value: Any, ttl: int = 300) -> None:
        """Set value in cache (both tiers)"""
        self.l1_cache.set(key, value, ttl)
        if self.l2_cache:
            self.l2_cache.set(key, value, ttl)

    def delete(self, key: str) -> None:
        """Delete key from all tiers"""
        self.l1_cache.delete(key)
        if self.l2_cache:
            self.l2_cache.delete(key)

    def clear(self) -> None:
        """Clear all tiers"""
        self.l1_cache.clear()
        if self.l2_cache:
            self.l2_cache.clear()

    def get_stats(self) -> dict:
        """Get cache statistics"""
        stats = {
            "l1": self.l1_cache.get_stats(),
            "l2": {"enabled": self.l2_cache is not None},
            "hits": {
                "l1": self.stats["l1_hits"],
                "l2": self.stats["l2_hits"],
                "total": self.stats["l1_hits"] + self.stats["l2_hits"],
            },
            "misses": self.stats["misses"],
            "hit_rate": 0.0,
        }

        total_requests = (
            self.stats["l1_hits"] + self.stats["l2_hits"] + self.stats["misses"]
        )
        if total_requests > 0:
            stats["hit_rate"] = (
                (self.stats["l1_hits"] + self.stats["l2_hits"]) / total_requests
            ) * 100

        return stats


def cache_key(*args, **kwargs) -> str:
    """Generate cache key from function arguments"""
    key_parts = [str(arg) for arg in args]
    key_parts.extend(f"{k}={v}" for k, v in sorted(kwargs.items()))
    key_string = ":".join(key_parts)
    return hashlib.md5(key_string.encode()).hexdigest()


def cached(ttl: int = 300, key_prefix: str = ""):
    """Decorator for caching function results"""

    def decorator(func: Callable) -> Callable:
        @wraps(func)
        def wrapper(*args, **kwargs):
            # Skip first arg if it's self
            cache_args = args[1:] if args and hasattr(args[0], "__dict__") else args

            # Generate cache key
            key = f"{key_prefix}:{func.__name__}:{cache_key(*cache_args, **kwargs)}"

            # Try to get from cache
            if hasattr(wrapper, "_cache"):
                cached_value = wrapper._cache.get(key)
                if cached_value is not None:
                    return cached_value

            # Call function
            result = func(*args, **kwargs)

            # Store in cache
            if hasattr(wrapper, "_cache"):
                wrapper._cache.set(key, result, ttl)

            return result

        return wrapper

    return decorator


class CacheWarmer:
    """Proactively warm cache with common queries"""

    def __init__(self, cache: MultiTierCache, data_fetcher):
        self.cache = cache
        self.data_fetcher = data_fetcher

    def warm_latest_blocks(self, count: int = 10) -> None:
        """Warm cache with latest blocks"""
        logger.info(f"Warming cache with latest {count} blocks")

        for i in range(count):
            try:
                # This would fetch and cache the latest blocks
                pass
            except Exception as e:
                logger.error(f"Error warming block cache: {e}")

    def warm_popular_addresses(self, addresses: list[str]) -> None:
        """Warm cache with popular addresses"""
        logger.info(f"Warming cache with {len(addresses)} addresses")

        for address in addresses:
            try:
                # This would fetch and cache address data
                pass
            except Exception as e:
                logger.error(f"Error warming address cache: {e}")

    def warm_validators(self) -> None:
        """Warm cache with validator data"""
        logger.info("Warming validator cache")

        try:
            # This would fetch and cache validator data
            pass
        except Exception as e:
            logger.error(f"Error warming validator cache: {e}")
