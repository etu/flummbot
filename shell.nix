with (import <nixpkgs> {});

mkShell {
  buildInputs = [
    gnumake
    go
    gocode
  ];
}
