# zunproxy

zunproxy is an HTTP proxy server with advanced caching control capabilities. Using memcached as a caching layer, it achieves high availability while reducing backend server load.

## Key Features

### Basic Features
- Functions as a standard HTTP proxy
- Fast response caching using memcached
- Flexible TTL settings based on response codes

### Advanced Cache Control
- Two-tier cache control with `SoftTTL` and `HardTTL`
  - `SoftTTL`: Standard cache expiration
  - `HardTTL`: Final cache expiration (for backend failure fallback)
- Concurrent request optimization
  - Prevents duplicate requests during cache updates (solves the issue of multiple backend requests occurring between cache expiration and update)
  - Effectively controls backend load

### High Availability Features
- Cache update timeout control
  - Falls back to existing cache when backend response is slow
  - Minimizes user wait times
- Continued cache serving during backend failures
  - Serves last successful response within `HardTTL` period

### Operations & Debug Features
- Request/response file dump functionality
- Conditional routing control based on:
  - Hostname
  - HTTP headers
  - Cookies
  - Query parameters

## Setup

### Requirements
- memcached
- zunproxy.cue (configuration file)

### Starting the Server
Place zunproxy.cue in the current directory and start with:
```bash
zunproxy
```

## Detailed Operation

### Cache Flow

1. Upon Request Receipt
   - Check cache existence
   - Return cache immediately if within `SoftTTL`

2. When `SoftTTL` Expires
   - Forward request to backend
   - Return existing cache if no response within `NewResponseWaitLimit (20ms)`
   - Cache updates independently once backend request resolves

3. Concurrent Request Control
   - Queues duplicate requests during cache updates
   - Returns new response to all waiting requests upon update completion

4. `HardTTL` Expiration
   - Completely removes cache
   - Processes as new request

## Important Notes

- Proper memcached configuration is required
- `HardTTL` must be set longer than `SoftTTL`
- Debug logging may impact performance

## License

This project is licensed under the MIT License - see the [LICENSE] file for details.
