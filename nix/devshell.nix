{ pkgs, perSystem }:
perSystem.devshell.mkShell {
  packages = [
    # go
    pkgs.go

    # python
    pkgs.poetry
    pkgs.python312

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

    {
      # Configure Poetry to use the Python version from Nix
      name = "POETRY_VIRTUALENVS_PATH";
      value = "./.venv";
    }
    {
      # keep Poetry's virtual environments in the project
      name = "POETRY_VIRTUALENVS_IN_PROJECT";
      value = "true";
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
