{ config, lib, pkgs, ... }:

let
  cfg = config.services.flummbot;

  package = pkgs.buildGoModule (let
    version = "20201209";
  in {
    name = "flummbot-${version}";

    src = ./.;

    prePatch = ''
      substituteInPlace main.go --replace "%version%" "${version}"
    '';

    vendorSha256 = "sha256-VIUm+m43fpFhflXU22hUD9IJINYZM6VWq7dmSiBsBGc=";
  });
in
{
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
      default = package;
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
        ExecStart = "${cfg.package}/bin/flummbot --config ${cfg.stateDirectory}/flummbot.toml";
        WorkingDirectory = cfg.stateDirectory;
        Restart = "always";
      };
    };
  };
}
