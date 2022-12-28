# mastobot

CLI for Mastodon bots.

## Build

### Apple Silicon

```bash
$ brew install FiloSottile/musl-cross/musl-cross
$ CC=x86_64-linux-musl-gcc \
  CXX=x86_64-linux-musl-g++ \
  GOARCH=amd64 \
  GOOS=linux \
  CGO_ENABLED=1 \
  go build \
  -ldflags "-s -w -linkmode external -extldflags -static" \
  -o mastobot .
```
