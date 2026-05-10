# ["Build Your Own Redis" Challenge](https://codecrafters.io/challenges/redis)

A Redis-compatible TCP server written in Go, built as part of the CodeCrafters challenge.

## Project Structure

```
app/
  main.go               # Entry point
internal/
  server/
    server.go           # TCP listener, connection handler, command dispatcher
  store/
    store.go            # In-memory key/value store
  resp/
    parser.go           # RESP protocol parser
    writer.go           # RESP response writers
```

## Progress

[![progress-banner](https://backend.codecrafters.io/progress/redis/42cf7ef4-ee9a-43a5-8f6c-1d210d61af35)](https://app.codecrafters.io/users/codecrafters-bot?r=2qF)
