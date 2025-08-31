# Load Test Results - Before Fixes

**Test Date**: August 31, 2025  
**Test Duration**: ~2.5 minutes  
**Production Server**: https://lnk.avantifellows.org

## Test Configuration

- **Total Concurrent Users**: 1,000
- **URLs Tested**: 10 test links (mactest01-mactest10)
- **Concurrent Users per URL**: 100
- **Total Requests**: 10,000
- **Test Tool**: hey (HTTP load testing tool)

## Summary Results

| Metric | Value |
|--------|--------|
| **Success Rate** | **65%** ⚠️ |
| **Total Requests** | 10,000 |
| **Successful Redirects** | 6,533 |
| **404 Errors** | 3,351 |
| **500 Errors** | 96 |
| **Timeout Errors** | 20 |
| **Average Latency** | 2.4 seconds |
| **Average Throughput** | 17.5 req/sec per URL |

## Critical Issues Identified

1. **High 404 Error Rate (33.5%)** - Database lookup failures under load
2. **Server Errors (0.96%)** - 500 responses indicating server instability
3. **Poor Latency (2.4s avg)** - Unacceptable for a redirect service
4. **Low Throughput** - Only ~175 total req/sec across all URLs

## Detailed Per-URL Results

| URL | Req/sec | Avg Latency | 404s | 500s | Timeouts |
|-----|---------|-------------|------|------|----------|
| mactest01 | 17.66 | 2.41s | 320 | 7 | 2 |
| mactest02 | 17.41 | 2.40s | 305 | 14 | 2 |
| mactest03 | 17.47 | 2.41s | 322 | 14 | 2 |
| mactest04 | 17.23 | 2.39s | 307 | 9 | 2 |
| mactest05 | 17.48 | 2.26s | 347 | 3 | 2 |
| mactest06 | 17.41 | 2.19s | 374 | 7 | 2 |
| mactest07 | 17.63 | 2.59s | 227 | 14 | 2 |
| mactest08 | 17.47 | 2.79s | 241 | 14 | 2 |
| mactest09 | 18.37 | 1.70s | 534 | 8 | 2 |
| mactest10 | 17.19 | 2.44s | 374 | 6 | 2 |

## Performance Characteristics

- **Response Time Range**: 0.04s (fastest) to 9.97s (slowest)
- **Most responses**: 1-3 second range
- **Consistent timeouts**: 2 per URL suggests connection limits
- **Variable 404 rates**: 227-534 per URL (inconsistent database behavior)

## Conclusion

The current production server cannot handle 1,000 concurrent users effectively:
- Only 65% success rate is far below production standards
- Database appears to have contention issues causing high 404 rates
- Server resources are overwhelmed leading to 500 errors and timeouts
- Performance degrades significantly under concurrent load

**Recommendation**: Critical fixes needed before handling higher loads.