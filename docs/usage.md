# Usage

## Installation

Nix is the supported distribution method:

```console
nix run github:traP-jp/neoshowcase-cli -- version
```

The flake exposes `packages.<system>.default`, `apps.<system>.default`, `devShells.<system>.default`, and `checks.<system>.default`.

## Authentication and configuration

Set the gateway and trusted-proxy identity at runtime:

```console
export NEOSHOWCASE_ENDPOINT=https://showcase.example.com
export NEOSHOWCASE_USER=alice
export NEOSHOWCASE_AUTH_HEADER=X-Showcase-User # optional; this is the default
neoshowcase-cli app list
```

Equivalent global flags are `--endpoint`, `--user`, and `--auth-header`. Values are resolved in this order: command-line flag, environment variable, configuration file, default. The default configuration file is `$XDG_CONFIG_HOME/neoshowcase-cli/config.json` (or the platform user config directory), and `--config` selects another file. JSON and simple `key = "value"` files are supported. Example JSON:

```json
{
  "endpoint": "https://showcase.example.com",
  "user": "alice",
  "auth_header": "X-Showcase-User"
}
```

Run `neoshowcase-cli --help` or `neoshowcase-cli <command> --help` for available commands, options, output behavior, and operational safety requirements.

## Security

TLS certificate verification is enabled by default and Go's standard CA configuration, including `SSL_CERT_FILE`, is honored. The explicit `--insecure-skip-verify` option prints a warning and should only be used for controlled testing. Authentication identity is never intentionally printed.

Application and build logs can contain secrets. Treat captured text, JSON, and JSON Lines output as sensitive data. Nix evaluation and builds do not read the endpoint or authentication environment variables; they are consumed only when the CLI runs.
