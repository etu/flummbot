{ pkgs ? import <nixpkgs> {}, ... }:

let
  version = "20210117";
in pkgs.buildGoModule {
  pname = "flummbot";
  inherit version;

  src = ./.;

  prePatch = ''
    substituteInPlace main.go --replace "%version%" ${version}
  '';

  vendorSha256 = "sha256-VIUm+m43fpFhflXU22hUD9IJINYZM6VWq7dmSiBsBGc=";
}
