{ pkgs ? import <nixpkgs> { } }:
with pkgs;

mkShell {
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
}
