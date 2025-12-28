{
  description = "Watchcat build process";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        gitRev = builtins.substring 0 8 (self.rev or self.dirtyRev);
        version = gitRev;

        recursiveMergeAttrs = listOfAttrsets: pkgs.lib.fold (attrset: acc: pkgs.lib.recursiveUpdate attrset acc) { } listOfAttrsets;

        # parameters to pass to functions
        parameterSet = {
          inherit self system pkgs version;
        };

        package = pkgs.callPackage ./nixos/package.nix { };

      in
      {
        packages."nspawn2" = package;
      }

    );
}
