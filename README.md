# cinecut-go-platform

Shared Go plumbing for the CineCut services.

```
go get github.com/cinecut-org/cinecut-go-platform@v1.1.0
```

| | |
|---|---|
| `Getenv`, `MustGetenv` | configuration |
| `WriteError`, `WriteJSON`, `DecodeJSON` | the `{"error","message"}` body every service returns, JSON in and out |
| `Health`, `Serve` | `/healthz`, and an HTTP server with graceful shutdown on SIGTERM |
| `IdentityFrom`, `RequireUser`, `RequireRole` | the caller as the gateway described it (`X-User-Id`, `X-User-Roles`, `X-Viewer-Country`); 401 / 403 |
| `InternalOnly` | guards `/internal/**` with `X-Internal-Token` (an empty token refuses everything) |
| `OpenDB`, `Migrate` | pgx pool; numbered `*.sql` migrations, each applied once, under an advisory lock |
| `LocalTime`, `ParseLocalTime`, `Now` | times written like the monolith's Java `LocalDateTime` (UTC, no zone) |
| `Page`, `NewPage` | the monolith's paged response |

Releases are git tags (`v1.0.0`, `v1.1.0`, ...). Services upgrade when they choose to. **Keep this repository public**
so service builds (Docker, CI) can fetch it without credentials; if you make it private, every service build needs a
token and `GOPRIVATE=github.com/cinecut-org/*`.

**v1.1.0** must be tagged and pushed before flags-audit or the gateway can be built from GitHub.

Still to come from the old `libs/go`: job lock.
