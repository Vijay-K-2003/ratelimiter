### Commands
```
docker build -t ratelimiter .
```
```
docker run -p 8080:8080 -d ratelimiter
```

### Using Command line
```
go run ./cmd/main.go --listen :8080 --max-requests-per-minute 10
```
### Visit http://localhost:8080/

