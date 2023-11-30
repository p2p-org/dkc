{
  perSystem = {
    pkgs,
    self',
    inputs',
    ...
  }: let
    inherit (inputs'.ethereum-nix.packages) mcl bls;
    pname = "dkc";
    version = "1.0.0";
    dkc = pkgs.buildGoModule {
      inherit pname version;
      src = ../.;

      vendorHash = "sha256-U2C3yuC0KE6bd7xURjW7uA4wGlQ92BXRpfHuV3CBjzA=";

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
