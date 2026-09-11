{
  description = "Third-party operational CLI for NeoShowcase";

  inputs = {
    flake-parts.url = "github:hercules-ci/flake-parts";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    inputs@{
      flake-parts,
      nixpkgs,
      self,
      ...
    }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = [
        "aarch64-darwin"
        "aarch64-linux"
        "x86_64-linux"
      ];

      perSystem =
        { system, ... }:
        let
          pkgs = import nixpkgs { inherit system; };
          cli = pkgs.callPackage ./nix/package.nix { src = self; };
        in
        {
          packages.default = cli;

          apps.default = {
            type = "app";
            program = "${cli}/bin/neoshowcase-cli";
            meta.description = "Run the NeoShowcase CLI";
          };

          checks.default = cli;

          devShells.default = pkgs.mkShell {
            packages = with pkgs; [
              go_1_25
              gopls
              gotools
            ];
            GOTOOLCHAIN = "local";
          };

          formatter = pkgs.nixfmt;
        };
    };
}
