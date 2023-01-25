{
  description = "etu/flummbot";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, flake-utils, nixpkgs, ... }: flake-utils.lib.eachDefaultSystem (system: let
    pkgs = import nixpkgs { inherit system; };
  in {
    packages = flake-utils.lib.flattenTree {
      default = pkgs.buildGoModule (let
        version = "2.1.${nixpkgs.lib.substring 0 8 self.lastModifiedDate}.${self.shortRev or "dirty"}";
      in {
        pname = "flummbot";
        inherit version;

        src = self;

        prePatch = ''
          substituteInPlace main.go --replace "%version%" ${version}
        '';

        vendorSha256 = "0rq4dhh4lrmpmdbaacqrsqh0klhgaildpm2mgrhr2zipdvx2d1al";
      });
    };

    # Set up flake module
    nixosModules.default = { options, config, lib, pkgs, ... }: let
      cfg = config.services.flummbot;
    in {
      # Set up module options
      options.services.flummbot = {
        enable = lib.mkEnableOption "Small IRC bot in go used for my channels";

        user = lib.mkOption {
          type = lib.types.str;
          default = "flummbot";
          defaultText = "flummbot";
          description = "Username of user running the bot software";
        };

        group = lib.mkOption {
          type = lib.types.str;
          default = "flummbot";
          defaultText = "flummbot";
          description = "Group of user running the bot software";
        };

        package = lib.mkOption {
          type = lib.types.package;
          default = self.packages.${system}.default;
          defaultText = "The package to use";
        };

        stateDirectory = lib.mkOption {
          type = lib.types.str;
          default = "/var/lib/flummbot";
          defaultText = "/var/lib/flummbot";
          description = "State directory of flummbot";
        };
      };

      # Set up module implementation
      config = lib.mkIf cfg.enable {
        users.users."${cfg.user}" = {
          description = "System user for flummbot";
          isSystemUser = true;
          inherit (cfg) group;
        };
        users.groups."${cfg.group}" = {};

        systemd.tmpfiles.rules = [ "d ${cfg.stateDirectory} 0700 ${cfg.user} ${cfg.group} -" ];

        systemd.services.flummbot = {
          description = "flummbot irc bot";
          after = [ "network.target" "network-online.target" ];
          wantedBy = [ "multi-user.target" ];
          serviceConfig = {
            Type = "simple";
            User = cfg.user;
            Group = cfg.group;
            ExecReload = "${pkgs.coreutils}/bin/kill -SIGUSR1 $MAINPID";
            ExecStart = "${cfg.package}/bin/flummbot --config ${cfg.stateDirectory}/flummbot.toml";
            WorkingDirectory = cfg.stateDirectory;
            Restart = "always";
          };
        };
      };
    };
  });
}
