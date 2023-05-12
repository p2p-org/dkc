{inputs, ...}: {
  imports = [
    inputs.pre-commit-hooks-nix.flakeModule
  ];
  perSystem = _: {
    pre-commit.settings = {
      hooks = {
        alejandra.enable = true;
        deadnix.enable = true;
        statix.enable = true;
        #gofmt.enable = true;
        #gotest.enable = true;
      };
    };
  };
}
