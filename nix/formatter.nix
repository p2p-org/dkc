{inputs, ...}: {
  imports = [
    inputs.treefmt-nix.flakeModule
  ];

  perSystem = {
    config,
    pkgs,
    ...
  }: {
    treefmt.config = {
      projectRootFile = ".git/config";
      package = pkgs.treefmt;

      programs = {
        alejandra.enable = true;
        gofmt.enable = true;
      };
    };

    formatter = config.treefmt.build.wrapper;
  };
}
