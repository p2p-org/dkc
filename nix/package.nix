{
  perSystem = {
    pkgs,
    self',
    inputs',
    ...
  }: let
    inherit (inputs'.ethereum-nix.packages) mcl bls;
    dkc = pkgs.buildGoModule {
      pname = "dkc";
      version = "1.0.0";
      src = ../.;

      vendorHash = "sha256-EtGm+9jpGGB+/aUzIyFfe3ZbyhqliL3G9qJBf2nKseY=";

      buildInputs = [mcl bls];
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
