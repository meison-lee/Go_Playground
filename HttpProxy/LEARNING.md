# HTTP Proxy Server Learning Project

## Project Overview
To practice concurrency in Go, I'm building an HTTP Proxy Server. The proxy will handle requests from clients and distribute them to backend servers. Using goroutines to handle each request, I'll learn about:
- Concurrency
- Locks
- Nginx-like concepts

## Version 1 Features

### 1. Docker Setup
- Using docker-compose to build two backend servers on ports 8081 and 8082
- Implemented CompileDaemon for automatic container rebuilds on code updates

### 2. Monitoring
- Request and error recording capabilities
- Monitoring endpoints:
  - `/metrics`
  - `/requests`

### 3. Routing System
- Each prefix has its own routePool containing backends
- Implemented basic RoundRobin load balancing
- **Limitations**:
  - No handling for downed servers
  - No auto-scaling support

### 4. Performance
- Implementation of read/write mutexes for improved efficiency

### 5. Error Handling
- Comprehensive error handling system

### 6. Configuration
- User-configurable backend settings via config file
- No hard-coded backends

### 7. Testing
- Implementation of test suite