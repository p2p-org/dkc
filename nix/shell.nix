{inputs, ...}: {
  imports = [
    inputs.devshell.flakeModule
  ];
  perSystem = {
    pkgs,
    config,
    inputs',
    ...
  }: let
    inherit (pkgs) go go-outline golangci-lint gopkgs gopls gotools openssl;
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
      ];
      commands = [
        {
          category = "Tools";
          name = "fmt";
          help = "Format the source tree";
          command = "nix fmt";
        }
        {
          category = "Tools";
          name = "check";
          help = "Nix flake check";
          command = "nix flake check";
        }
      ];
      devshell.startup = {
        pre-commit.text = config.pre-commit.installationScript;
      };
    };
  };
}
