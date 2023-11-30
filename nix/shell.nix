{inputs, ...}: {
  imports = [
    inputs.devshell.flakeModule
  ];
  perSystem = {
    pkgs,
    inputs',
    ...
  }: let
    inherit (pkgs) go go-outline golangci-lint gopkgs gopls gotools openssl act;
    inherit (inputs'.ethereum-nix.packages) ethdo;
  in {
    devshells.default = {
      name = "dkc";
      packages = [
        go
        go-outline
        golangci-lint
        gopkgs
        gopls
        gotools
        openssl
        ethdo
        act
      ];
    };
  };
}
