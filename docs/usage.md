# Usage

## Installation

Nix is the supported distribution method:

```console
nix run github:traP-jp/neoshowcase-cli -- version
```

The flake exposes `packages.<system>.default`, `apps.<system>.default`, `devShells.<system>.default`, and `checks.<system>.default`.

## Authentication and endpoint

Set the trusted-proxy identity at runtime:

```console
export NEOSHOWCASE_USER=alice
export NEOSHOWCASE_AUTH_HEADER=X-Showcase-User # optional; this is the default
neoshowcase-cli app list
```

The endpoint is selected with `--endpoint` and defaults to `https://ns.trap.jp`. `NEOSHOWCASE_USER` is required for commands that access the API. `NEOSHOWCASE_AUTH_HEADER` is optional and defaults to `X-Showcase-User`.

Run `neoshowcase-cli --help` or `neoshowcase-cli <command> --help` for available commands, options, output behavior, and operational safety requirements.

## Security

TLS certificate verification is enabled by default and Go's standard CA configuration, including `SSL_CERT_FILE`, is honored. The explicit `--insecure-skip-verify` option prints a warning and should only be used for controlled testing. Authentication identity is never intentionally printed.

Application and build logs can contain secrets. Treat captured text, JSON, and JSON Lines output as sensitive data. Nix evaluation and builds do not read the authentication environment variables; they are consumed only when the CLI runs.
