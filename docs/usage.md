# Usage

## Installation

Nix is the supported distribution method:

```console
nix run github:traP-jp/neoshowcase-cli -- version
```

The flake exposes `packages.<system>.default`, `apps.<system>.default`, `devShells.<system>.default`, and `checks.<system>.default`.

## Authentication and endpoint

Copy the value of the `Cookie` request header from an authenticated NeoShowcase browser session. Do not include the `Cookie:` prefix or attributes from a `Set-Cookie` response header.

```console
export NEOSHOWCASE_SESSION_COOKIE='cookie_name=cookie_value; another_cookie=another_value'
neoshowcase-cli app list
```

The endpoint is selected with `--endpoint` and defaults to `https://ns.trap.jp`. `NEOSHOWCASE_SESSION_COOKIE` is required for commands that access the API.

Run `neoshowcase-cli --help` or `neoshowcase-cli <command> --help` for available commands, options, output behavior, and operational safety requirements.

## Security

TLS certificate verification is always enabled, and Go's standard CA configuration, including `SSL_CERT_FILE`, is honored. The session cookie is never intentionally printed.

Application and build logs can contain secrets. Treat captured text, JSON, and JSON Lines output as sensitive data. Nix evaluation and builds do not read the authentication environment variables; they are consumed only when the CLI runs.
