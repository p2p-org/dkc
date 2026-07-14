{
  perSystem = {
    pkgs,
    self',
    inputs',
    ...
  }: let
    inherit (inputs'.ethereum-nix.packages) mcl bls;
    pname = "dkc";
    version = "1.1.0";
    dkc = pkgs.buildGoModule {
      inherit pname version;
      src = ../.;

      vendorHash = "sha256-6W+hUANdSqDjCAaucIhRPGQQzUIuKK+Fihbp3KG7osw=";

      buildInputs = [mcl bls];

      env.CGO_LDFLAGS = "-lmcl";

      ldflags = [
        "-s"
        "-w"
        "-X github.com/p2p-org/dkc/cmd.ReleaseVersion=${version}"
      ];
    };
  in {
    packages.dkc = dkc;
    packages.default = self'.packages.dkc;
    apps.dkc = {
      type = "app";
      program = "${self'.packages.default}/bin/dkc";
    };
    apps.default = self'.apps.dkc;
  };
}
