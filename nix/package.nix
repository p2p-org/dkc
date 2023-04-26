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

      vendorHash = "sha256-FN8+W+OZ/XGsO0Kt0PJZoT+56dxCSJLGPX1KK6E4ozc=";

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
