{ pkgs, nix-filter, nhost }:
let
  name = "graphql";
  submodule = "lib/${name}";
  description = "Common graphql functionality";
  version = pkgs.lib.fileContents ./VERSION;

  # source files needed for the build
  src = nix-filter.lib.filter {
    root = ../..;
    include = with nix-filter.lib;[
      ".golangci.yaml"
      "go.mod"
      "go.sum"
      (inDirectory "vendor")
      isDirectory
      (and
        (inDirectory submodule)
        (matchExt "go")
      )
      (and
        (inDirectory "lib/consoleNextClient")
        (matchExt "go")
      )
      (and
        (inDirectory "lib/tracing")
        (matchExt "go")
      )
    ];
  };

  checkDeps = with pkgs; [
    mockgen
  ];

  buildInputs = with pkgs; [
  ];

  nativeBuildInputs = with pkgs; [
  ];

  tags = [ ];

  ldflags = [
  ];
in
rec{
  inherit name description version;

  check = nhost.go.check {
    inherit src submodule ldflags tags checkDeps buildInputs nativeBuildInputs;
  };

  devShell = nhost.go.devShell {
    buildInputs = with pkgs; [
    ] ++ checkDeps ++ buildInputs ++ nativeBuildInputs;
  };

  package = nhost.go.package {
    inherit name description version src submodule ldflags buildInputs nativeBuildInputs;
  };
}
