# mastobot

CLI for Mastodon bots.

Application client credentials and access tokens (but not account
username/password) are stored in plaintext in a sqlite database. Access to this
database file should be protected.

## Usage

1. Create a new account (varies by instance)

2. Register application with that instance

```bash
$ mastobot app register --instance <instance> --name
```

3. Get an access token for the account

```bash
$ mastobot app token renew --instance <instance> --name <appName> --email <email> --password <password>
```

4. Send a test toot

```bash
$ mastobot app register --instance <instance> --name <appName> --visibility public 'Hello from mastobot!'
```

## Build

```bash
$ devbox run build
```

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
