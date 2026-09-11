# Vendored NeoShowcase API

The generated Connect RPC client in `internal/api` is generated from
`neoshowcase/protobuf/gateway.proto` at NeoShowcase revision
`16eda27a8cda8858811406411bcc7f2f508e9efc`.

When updating the revision, regenerate both protobuf and Connect bindings and
verify the official dashboard polling interval. The CLI interval is 1.1 times
that value; this revision polls application and build state every 10 seconds,
so the CLI uses 11 seconds.
