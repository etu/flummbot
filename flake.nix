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

        vendorSha256 = "sha256-VIUm+m43fpFhflXU22hUD9IJINYZM6VWq7dmSiBsBGc=";
      };
    };

    # Set it as default package for all architectures
    defaultPackage = forAllSystems (system: (import nixpkgs {
      inherit system;
      overlays = [ self.overlay ];
    }).flummbot);

    # Set up flake module
    nixosModule = { inputs, options, modulesPath, config, lib }: let
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
          default = self.defaultPackage.x86_64-linux;
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
    };
  };
}
