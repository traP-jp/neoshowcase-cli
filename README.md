# neoshowcase-cli

`neoshowcase-cli` is a third-party operational client for NeoShowcase. It reads
application and build state, follows logs, waits for builds, and performs a
small set of explicitly guarded runtime operations. Declarative configuration
of repositories, applications, and environment variables remains outside its
scope.

## Installation

Nix is the supported distribution method:

```console
nix run github:traP-jp/neoshowcase-cli -- version
```

The flake exposes `packages.<system>.default`, `apps.<system>.default`,
`devShells.<system>.default`, and `checks.<system>.default`.

## Authentication and configuration

Set the gateway and trusted-proxy identity at runtime:

```console
export NEOSHOWCASE_ENDPOINT=https://showcase.example.com
export NEOSHOWCASE_USER=alice
export NEOSHOWCASE_AUTH_HEADER=X-Showcase-User # optional; this is the default
neoshowcase-cli app list
```

Equivalent global flags are `--endpoint`, `--user`, and `--auth-header`. Values
are resolved in this order: command-line flag, environment variable,
configuration file, default. The default configuration file is
`$XDG_CONFIG_HOME/neoshowcase-cli/config.json` (or the platform user config
directory), and `--config` selects another file. JSON and simple `key =
"value"` files are supported. Example JSON:

```json
{
  "endpoint": "https://showcase.example.com",
  "user": "alice",
  "auth_header": "X-Showcase-User"
}
```

## Commands

```text
neoshowcase-cli app list
neoshowcase-cli app get <application>
neoshowcase-cli app logs <application> [--follow] [--tail N] [--since TIME]
neoshowcase-cli app start|stop|restart <application> --allow-mutable-operation
neoshowcase-cli app rebuild <application> [--commit SHA] [--wait] [--logs] --allow-mutable-operation

neoshowcase-cli build list [application] [--page PAGE] [--limit N]
neoshowcase-cli build get <build-id>
neoshowcase-cli build logs <build-id> [--follow]
neoshowcase-cli build wait <build-id> [--logs]
neoshowcase-cli build watch <application> --commit SHA [--logs]
neoshowcase-cli build retry <build-id> [--wait] [--logs] --allow-mutable-operation
neoshowcase-cli build cancel <build-id> --allow-mutable-operation
```

Application arguments accept either an ID or a unique exact name. Waiting and
streaming commands default to a whole-command timeout of 10 minutes and accept
`--timeout`. On rebuild and retry, `--logs` implies `--wait`. Build polling
occurs immediately and then every 11 seconds.

All information and monitoring commands support `--output text|json|jsonl`.
Stream records use JSON Lines in either machine-readable streaming mode.
Diagnostics are written to stderr. Times are emitted in UTC RFC 3339 form.
Diagnostic verbosity is controlled with `--log-level`; the default is `warn`.

Every state-changing invocation requires `--allow-mutable-operation`; it cannot
be enabled through an environment variable or configuration file. Restart uses
a non-atomic state check followed by NeoShowcase's `StartApplication` call, so a
concurrent server-side state change can still race that check.

## Security

TLS certificate verification is enabled by default and Go's standard CA
configuration, including `SSL_CERT_FILE`, is honored. The explicit
`--insecure-skip-verify` option prints a warning and should only be used for
controlled testing. Authentication identity is never intentionally printed.

Application and build logs can contain secrets. Treat captured text, JSON, and
JSON Lines output as sensitive data. Nix evaluation and builds do not read the
endpoint or authentication environment variables; they are consumed only when
the CLI runs.
