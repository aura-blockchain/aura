# Block Explorer Cache System

Multi-tier caching implementation with Redis and in-memory fallback.

## Features

- **Multi-tier caching**: L1 (Memory) + L2 (Redis)
- **Automatic fallback**: Gracefully falls back to MemoryCache if Redis unavailable
- **LRU eviction**: Memory cache uses least-recently-used eviction
- **TTL support**: Automatic expiration of cached entries
- **Key prefixing**: Namespace isolation for different cache instances
- **Statistics**: Detailed cache hit/miss tracking

## Quick Start

### Basic Usage

```python
from cache import RedisCache, MultiTierCache

# Redis cache with automatic fallback
cache = RedisCache(redis_url="redis://localhost:6379/0")
cache.set("key", {"data": "value"}, ttl=300)
value = cache.get("key")

# Multi-tier cache (recommended)
multi = MultiTierCache()
multi.set("key", "value", ttl=300)
value = multi.get("key")
```

### Configuration

Redis URL can be configured via environment variable:
```bash
export REDIS_URL="redis://localhost:6379/0"
```

Or passed directly:
```python
cache = RedisCache(redis_url="redis://myhost:6379/1", key_prefix="app:")
```

## Components

### MemoryCache
In-memory LRU cache with TTL support.

```python
from cache import MemoryCache

cache = MemoryCache(max_size=1000)
cache.set("key", "value", ttl=60)
value = cache.get("key")
```

### RedisCache
Redis-backed cache with automatic fallback to MemoryCache.

```python
from cache import RedisCache

cache = RedisCache(
    redis_url="redis://localhost:6379/0",
    key_prefix="explorer:",
)

# Test connection
if cache.test_connection():
    print("Redis connected")

# Operations
cache.set("block:1234", {"height": 1234}, ttl=300)
block = cache.get("block:1234")
cache.delete("block:1234")
cache.clear()  # Clear all keys with prefix

# Statistics
stats = cache.get_stats()
print(f"Hit rate: {stats.get('keyspace_hits', 0)}")
```

### MultiTierCache
Combined L1 (memory) + L2 (Redis) cache with automatic promotion.

```python
from cache import MultiTierCache, RedisCache

redis = RedisCache(redis_url="redis://localhost:6379/0")
cache = MultiTierCache(redis_cache=redis)

cache.set("key", "value")
value = cache.get("key")  # Checks L1, then L2, promotes to L1

stats = cache.get_stats()
print(f"L1 hits: {stats['hits']['l1']}")
print(f"L2 hits: {stats['hits']['l2']}")
print(f"Hit rate: {stats['hit_rate']:.2f}%")
```

## Fallback Behavior

RedisCache automatically falls back to MemoryCache when:
- Redis server is unavailable
- redis-py package not installed
- Connection errors occur
- Serialization errors happen

Operations continue to work seamlessly via the fallback cache.

## Testing

Run automated tests:
```bash
cd explorer
python3 -m pytest test_cache.py -v
```

Run manual integration tests:
```bash
python3 test_redis_manual.py
```

## Dependencies

Required:
- Python 3.12+

Optional (for Redis support):
- redis==5.0.1

Install:
```bash
pip3 install redis==5.0.1
```

## Performance

- **L1 (Memory)**: Sub-microsecond access
- **L2 (Redis)**: ~1ms local, faster with pipelining
- **Fallback**: Automatic, no manual intervention

## Best Practices

1. Use MultiTierCache for production workloads
2. Set appropriate TTL values (default: 300s)
3. Use key prefixes to avoid collisions
4. Monitor cache stats for optimization
5. Clear cache on schema changes

## Environment Variables

- `REDIS_URL`: Redis connection URL (default: redis://localhost:6379/0)

## Example: Block Explorer Integration

```python
from cache import MultiTierCache, RedisCache
import os

# Initialize cache
redis_cache = RedisCache(
    redis_url=os.getenv("REDIS_URL", "redis://localhost:6379/0"),
    key_prefix="explorer:",
)
cache = MultiTierCache(redis_cache=redis_cache)

# Cache block data
def get_block(height):
    key = f"block:{height}"
    block = cache.get(key)

    if block is None:
        # Fetch from chain
        block = fetch_block_from_chain(height)
        cache.set(key, block, ttl=600)

    return block
```
