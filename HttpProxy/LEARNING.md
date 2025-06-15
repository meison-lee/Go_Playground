To practice concurrency in Go, ChatGPT told me that I can try to build a Http Proxy Server, the proxy will handle the requests from client and distributes them to backend server, goroutine will handle each requests and I can learn about concurrency, lock and knowledge about Nginx.


Version 1 feature:

1. Using docker-compose to build-up two backend server on :8081, :8082. With the CompileDaemon, I can rebuild the container as soon as I updated the code.
2. The proxy can record the requests and errors, you can monitor them in /metrics or /requests
3. Each prefix has its own routePool contain backends ready to serve the request, I also save an index to implement RoundRobin roughly(I haven't implement situation when the backend server is downed or auto-scaled...)
4. It should use read, write mutex to improve the efficiency.
5. It should read the config file from user, instead of hard-coding backends.