{ pkgs ? import <nixpkgs> {}, ... }:

let
  version = "20210130";
in pkgs.buildGoModule {
  pname = "flummbot";
  inherit version;

  src = ./.;

  prePatch = ''
    substituteInPlace main.go --replace "%version%" ${version}
  '';

  vendorSha256 = "0rq4dhh4lrmpmdbaacqrsqh0klhgaildpm2mgrhr2zipdvx2d1al";
}
