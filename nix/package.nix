{
  perSystem = {
    pkgs,
    self',
    inputs',
    ...
  }: let
    inherit (inputs'.ethereum-nix.packages) mcl bls;
    pname = "dkc";
    version = "0.1.1";
    dkc = pkgs.buildGoModule {
      inherit pname version;
      src = ../.;

      vendorHash = "sha256-EtGm+9jpGGB+/aUzIyFfe3ZbyhqliL3G9qJBf2nKseY=";

      buildInputs = [mcl bls];

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
