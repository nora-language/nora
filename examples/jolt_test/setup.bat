@echo off
setlocal

echo Setting up JoltPhysics Example...

if not exist JoltPhysics (
    echo Cloning JoltPhysics...
    git clone https://github.com/jrouwe/JoltPhysics.git JoltPhysics
)

cd JoltPhysics
echo Checking out stable version 5.5.0...
git checkout v5.5.0

cd Build
echo Configuring CMake for JoltPhysics...
cmake -B . -A x64

echo Building Release static library (This will take a few minutes)...
cmake --build . --config Release

echo Copying Jolt.lib for Nora to link...
cd ..\..
if not exist "lib" mkdir "lib"
copy JoltPhysics\Build\Release\Jolt.lib "lib\"

clang++ -std=c++17 -O3 -DNDEBUG -mavx2 -mfma -mlzcnt -mbmi -mf16c -D JPH_PROFILE_ENABLED -D JPH_DEBUG_RENDERER -D JPH_OBJECT_STREAM -D JPH_FLOATING_POINT_EXCEPTIONS_ENABLED -c src\wrapper.cpp -I JoltPhysics -o lib\wrapper.obj

echo Setup complete!
echo You can now run the example using: Nora run .
