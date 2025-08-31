# Redis Implementation TODO

## Performance Analysis Summary

### Current Bottleneck
- 50K concurrent users overwhelm SQLite with database queries
- Each redirect requires SQLite lookup: `SELECT original_url FROM link_mappings WHERE short_code = ?`
- t3.medium (2 vCPU) insufficient for 50K concurrent SQLite operations

### With Redis Caching
- **95% cache hit ratio** for active session links
- **CPU usage**: 20-30% during 50K peaks (vs 60-80% without Redis)
- **Memory usage**: 350MB total (150MB Go + 200MB Redis)
- **Performance**: t3.medium becomes sufficient for 50K concurrent users

## Redis Configuration

### Memory Management
```redis
maxmemory 200mb
maxmemory-policy allkeys-lru
```

### Expected Capacity
- **~50,000 cached links** (4KB average per link)
- **LRU eviction**: Inactive session links automatically removed
- **Active links**: Stay cached as long as they're accessed

## Implementation Tasks

### 1. Infrastructure Setup
- [ ] Add Redis installation to `deploy.sh`
- [ ] Configure Redis memory limits and LRU eviction
- [ ] Update systemd service dependencies

### 2. Code Changes
- [ ] Add `github.com/redis/go-redis/v9` to `go.mod`
- [ ] Create Redis client in main.go
- [ ] Modify `RedirectURL` handler for cache-first lookup
- [ ] Add cache write on SQLite miss
- [ ] Add cache invalidation for link deletion (if needed)

### 3. Environment Configuration
- [ ] Add Redis connection settings to `.env.production`
- [ ] Configure Redis host/port (default: localhost:6379)

## Usage Pattern Optimization

### Session-Based Access Pattern
- Links active for ~1 day (session duration)
- Monthly peaks of 50K concurrent users
- Normal load: <100 concurrent users (1-5% CPU)

### LRU Strategy Benefits
- No arbitrary TTL expiration
- Active session links stay cached indefinitely
- Memory automatically adjusts to usage patterns
- Inactive links evicted only when memory needed

## Resource Utilization

### Normal Operation (99% of time)
- **CPU**: 1-5%
- **Memory**: 200MB (Go + minimal Redis cache)
- **Instance**: t3.medium accumulates CPU credits

### Peak Load (1% of time, monthly)
- **CPU**: 20-30% (well within t3.medium burst capacity)
- **Memory**: 350MB (plenty of headroom in 4GB)
- **Performance**: Smooth operation with Redis caching

## Implementation Estimate
- **Effort**: ~30 minutes
- **Risk**: Low (Redis is stable, Go client is mature)
- **Impact**: Enables 50K concurrent user support on existing infrastructure