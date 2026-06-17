@echo off
setlocal

echo Setting up wgpu-native...
mkdir lib\wgpu 2>nul
cd lib\wgpu

if not exist "wgpu-windows-x86_64-msvc-release.zip" (
    echo Downloading wgpu-native...
    curl -L -o wgpu-windows-x86_64-msvc-release.zip https://github.com/gfx-rs/wgpu-native/releases/latest/download/wgpu-windows-x86_64-msvc-release.zip
)

echo Extracting wgpu-native...
tar -xf wgpu-windows-x86_64-msvc-release.zip
copy wgpu.dll ..\..\wgpu.dll

echo Done!
cd ..\..
echo Setup complete. You can now build the project.
