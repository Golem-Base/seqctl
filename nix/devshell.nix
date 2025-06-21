{ pkgs, perSystem }:
perSystem.devshell.mkShell {
  packages = [
    # go
    pkgs.go
    pkgs.goreleaser
    pkgs.revive
    pkgs.templ

    # k8s
    pkgs.k9s
    pkgs.kind
    pkgs.kubebuilder
    pkgs.kubectl
    pkgs.kubelogin-oidc
    pkgs.kubernetes-code-generator
    pkgs.kubernetes-controller-tools
    pkgs.operator-sdk

    # other
    perSystem.self.formatter
    pkgs.go-swag
    pkgs.just
  ];

  env = [
    {
      name = "NIX_PATH";
      value = "nixpkgs=${toString pkgs.path}";
    }
    {
      name = "NIX_DIR";
      eval = "$PRJ_ROOT/nix";
    }
  ];

  commands = [
    {
      name = "k";
      category = "ops";
      help = "Shorter alias for kubectl";
      command = ''${pkgs.kubectl}/bin/kubectl "$@"'';
    }
    {
      name = "kvs";
      category = "Ops";
      help = "kubectl view-secret alias";
      command = ''${pkgs.kubectl-view-secret}/bin/kubectl-view-secret "$@"'';
    }
    {
      name = "kns";
      category = "ops";
      help = "Switch kubernetes namespaces";
      command = ''kubens "$@"'';
    }
  ];
}
