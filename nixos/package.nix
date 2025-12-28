# nix-build -E 'with import <nixpkgs> {}; callPackage ./nixos/package.nix {}'
{ lib
, buildGo124Module
, fetchgit
, tag ? "0.6.2"
, hash ? ""
}:

buildGo124Module rec {
  pname = "nomad-driver-nspawn2";
  version = "${tag}";

  src = fetchgit {
    url = "https://codeberg.org/stackshadow/nspawn2";
    rev = "refs/tags/v${tag}";
    hash = hash;
  };

  # src = /mnt/synced/develop/nomad/nspawn-reduced;

  vendorHash = "sha256-HLXHegOk33jUkrjBsBpzkHbs8/CbGqBiJomkpXsXNKc=";

  subPackages = [ "." ];

  # some tests require a running podman service
  doCheck = false;

  # BUILDARGS := build -mod=vendor -a -v -ldflags '-extldflags "-static" -X github.com/JanMa/nomad-driver-nspawn/nspawn.pluginVersion=${VERSION}' -o $(BINARY)

  ldflags = [
    "-X github.com/stackshadow/nspawn2/nspawndriver.pluginVersion=${tag}"
  ];

  meta = with lib; {
    homepage = "https://codeberg.org/stackshadow/nspawn2";
    description = "nspawn task driver for Nomad";
    mainProgram = "nspawn2";
    platforms = platforms.linux;
    license = licenses.mpl20;
    maintainers = with maintainers; [ ];
  };
}
