{
  perSystem = {
    pkgs,
    config,
    inputs',
    ...
  }: let
    inherit (pkgs) mkShell;
  in {
    devShells.default = mkShell {
      name = "dkc";
      inputsFrom = [
        config.flake-root.devShell
        config.mission-control.devShell
        #config.pre-commit.devShell
      ];

      packages = builtins.attrValues {
        inherit
          (pkgs)
          go
          go-outline
          golangci-lint
          gopkgs
          gopls
          gotools
          openssl
          ;
        inherit (inputs'.ethereum-nix.packages) ethdo;
      };
      shellHook = ''
        ${config.pre-commit.installationScript}
      '';
    };
  };
}
