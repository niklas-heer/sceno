{
  lib,
  buildGo127Module,
  versionCheckHook,
}:

buildGo127Module {
  pname = "sceno";
  version = lib.trim (builtins.readFile ../internal/version/VERSION);

  src = lib.cleanSource ../.;

  vendorHash = "sha256-o2v5I35gGZqIxrZoVoNspKNlatrkJpn/lL4yx7Co2w0=";

  subPackages = [ "cmd/sceno" ];
  env.CGO_ENABLED = 0;
  ldflags = [
    "-s"
    "-w"
  ];

  # CI runs the full suite; the package build only proves the binary starts.
  doCheck = false;
  doInstallCheck = true;
  nativeInstallCheckInputs = [ versionCheckHook ];

  meta = {
    description = "Declarative architecture diagrams in KDL";
    homepage = "https://github.com/niklas-heer/sceno";
    license = lib.licenses.mit;
    mainProgram = "sceno";
    platforms = lib.platforms.unix;
  };
}
