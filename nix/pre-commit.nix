{ inputs, ... }: {
  imports = [
    inputs.pre-commit-hooks-nix.flakeModule
  ];
  perSystem = { config, ... }: {
    pre-commit.settings = {
      hooks = {
        # Same formatter as `nix fmt` (treefmt: nixfmt + gofmt), so the hook
        # can never fight the editor or the flake formatter
        treefmt = {
          enable = true;
          package = config.treefmt.build.wrapper;
        };
        deadnix.enable = true;
        statix.enable = true;
        #gotest.enable = true;
      };
    };
  };
}
