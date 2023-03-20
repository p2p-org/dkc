{
  description = "Dirk Key Converter";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    ethereum-nix = {
      url = "github:nix-community/ethereum.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, ethereum-nix }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };

      dkc = pkgs.buildGoModule {
        pname = "dkc";
        version = "0.1.0";
        src = ./.;
      };
    in {
      packages."x86_64-linux".key-converter = dkc;
      packages."x86_64-linux".default = self.packages."${system}".dkc;
      apps."x86_64-linux".dkc = {
        type = "app";
        program = "${self.packages.x86_64-linux.default}/bin/dkc";
      };
      apps."x86_64-linux".default = self.apps."${system}".dkc;

      devShells."x86_64-linux".default = pkgs.mkShell {
        buildInputs = with pkgs; [
          ethereum-nix.packages.x86_64-linux.ethdo
          go
          go-outline
          golangci-lint
          gopkgs
          gopls
          gotools
          openssl
        ];
      };
    };
}
