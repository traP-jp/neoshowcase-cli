{
  buildGoModule,
  lib,
  src,
}:

buildGoModule {
  pname = "neoshowcase-cli";
  version = "0.1.0";
  inherit src;

  vendorHash = "sha256-9oCFmML2HP6KfXXlsr/iCI5lLP01fK9OR8WlLWWwh3w=";
  subPackages = [ "cmd/neoshowcase-cli" ];
  ldflags = [
    "-s"
    "-w"
    "-X main.version=0.1.0"
  ];

  meta = {
    description = "Third-party operational CLI for NeoShowcase";
    homepage = "https://github.com/traP-jp/neoshowcase-cli";
    license = lib.licenses.mit;
    mainProgram = "neoshowcase-cli";
  };
}
