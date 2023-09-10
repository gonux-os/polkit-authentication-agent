{
  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
  };
  outputs = { self, nixpkgs, ... }:
    let
      pkgs = nixpkgs.legacyPackages.x86_64-linux;
    in
    with pkgs;
  {

    packages.x86_64-linux.default = buildGoModule {
      name = "polkit-authentication-agent";
      version = "0.0.1";
      src = ./.;
      vendorHash = "sha256-0/Y5cmZkXfISlYiPqkY/XhWHiKq//xXhVBAk2TE5ISE=";
      buildInputs = with pkgs; [
        dbus
        wayland
        vulkan-headers
        libxkbcommon
        xorg.libX11
        xorg.libX11.dev
        xorg.libXi
        xorg.libXinerama
        xorg.libXrandr
        xorg.libXxf86vm
        xorg.libXcursor
        xorg.libXfixes
        libGL
      ];
      nativeBuildInputs = with pkgs; [
        pkgconfig
      ];
      inherit (pkgs) dbus;
    };

  };
}
