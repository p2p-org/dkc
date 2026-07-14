{inputs, ...}: {
  imports = [
    inputs.devshell.flakeModule
  ];
  perSystem = {
    pkgs,
    inputs',
    config,
    ...
  }: let
    inherit
      (pkgs)
      go
      go-outline
      golangci-lint
      gopkgs
      gopls
      gotools
      openssl
      act
      gmp
      gcc
      pinact
      zizmor
      ;
    inherit (inputs'.ethereum-nix.packages) ethdo;
  in {
    devshells.default = {
      name = "dkc";
      env = [
        {
          name = "CGO_ENABLED";
          eval = "1";
        }
      ];
      packages = [
        pinact
        go
        gcc
        go-outline
        golangci-lint
        gopkgs
        gopls
        gotools
        openssl
        ethdo
        act
        gmp
        zizmor
      ];
      devshell.startup = {
        pre-commit.text = config.pre-commit.installationScript;
      };
    };
  };
}
