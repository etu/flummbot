{
  description = "etu/flummbot";

  outputs = { self, nixpkgs }: let
    supportedSystems = [ "x86_64-linux" ];
    forAllSystems = f: nixpkgs.lib.genAttrs supportedSystems (system: f system);
    version = "2.0.${nixpkgs.lib.substring 0 8 self.lastModifiedDate}.${self.shortRev or "dirty"}";
  in {
    # Set up an overlay with the flummbot package
    overlay = final: prev: {
      flummbot = final.pkgs.buildGoModule {
        name = "flummbot-${version}";
        src = self;

        prePatch = ''
          substituteInPlace main.go --replace "%version%" "${version}"
        '';

        vendorSha256 = "sha256-FjvTVVPO7M74dcqJCP6eaPilhLAA2lMFEUyDtoBGgzI=";
      };
    };

    # Set it as default package for all architectures
    defaultPackage = forAllSystems (system: (import nixpkgs {
      inherit system;
      overlays = [ self.overlay ];
    }).flummbot);
  };
}
