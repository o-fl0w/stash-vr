{
  description = "stash-vr (gomod2nix)";

  inputs = {
    nixpkgs.url   = "github:NixOS/nixpkgs/nixos-25.05";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix.url = "github:nix-community/gomod2nix";
  };

  outputs = { self, nixpkgs, flake-utils, gomod2nix }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ gomod2nix.overlays.default ];
        };
      in {
        packages.default = pkgs.buildGoApplication {
          pname = "stash-vr";
          version = "0.9.6";
          src = ./.;
          modules = ./gomod2nix.toml;
          #subPackages = [ "cmd/stash-vr" ];
        };
      });
}

