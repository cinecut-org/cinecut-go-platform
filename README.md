# cinecut-go-platform

Shared Go plumbing for the CineCut services: config, the JSON error body, health check, identity headers.

```
go get github.com/cinecut-org/go-platform@v1.0.0
```

Releases are git tags (`v1.0.0`, `v1.1.0`, ...). Services upgrade when they choose to. **Keep this repository public**
so service builds (Docker, CI) can fetch it without credentials; if you make it private, every service build needs a
token and `GOPRIVATE=github.com/cinecut-org/*`.

Still to come from the old `libs/go`: DB pool and job lock.
