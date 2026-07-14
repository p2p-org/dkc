{ self, ... }: {
  perSystem =
    {
      pkgs,
      self',
      inputs',
      ...
    }:
    let
      inherit (inputs'.ethereum-nix.packages) mcl bls;
      pname = "dkc";
      baseVersion = "1.2.0";
      # Pure flake eval cannot see git tags, only the checked-out revision:
      # suffix the base version with it so binaries are traceable to a commit.
      # shortRev exists for clean trees, dirtyShortRev for dirty ones.
      version = "${baseVersion}+${self.shortRev or self.dirtyShortRev or "unknown"}";
      dkc = pkgs.buildGoModule {
        inherit pname version;
        src = ../.;

        vendorHash = "sha256-6W+hUANdSqDjCAaucIhRPGQQzUIuKK+Fihbp3KG7osw=";

        buildInputs = [
          mcl
          bls
        ];

        env.CGO_LDFLAGS = "-lmcl";

        ldflags = [
          "-s"
          "-w"
          "-X github.com/p2p-org/dkc/cmd.ReleaseVersion=${version}"
        ];
      };
    in
    {
      packages.dkc = dkc;
      packages.default = self'.packages.dkc;
      apps.dkc = {
        type = "app";
        program = "${self'.packages.default}/bin/dkc";
      };
      apps.default = self'.apps.dkc;
    };
}
