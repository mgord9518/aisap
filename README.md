# aisap

<p align="center"><img src="resources/aisap.svg" width=256 height="256"/></p>

AppImage SAndboxing Project (pronounced /eɪsæp/): a tool to help sandbox
AppImages through bwrap.

**EARLY DEVELOPMENT!** Many parts of this are subject to change and should
be expected to until it reaches a more stable form.

## What is it?
aisap intends to be a simple way to implement Android/Flatpak style sandboxing
with AppImages. It has a profile system, but it intends to keep it as basic as
possible, making it easier to understand what a program actually requires to
run without dealing with the hassle of individually cherry-picking files.

It currently has a basic re-implementaion of the go-appimage API, so modifying
existing Go programs to include sandboxing should be fairly painless

## Using aisap:
 1. [aisap cli](docs/aisap.1.md)
 2. [aisap Go implementation](docs/aisap-go.3.md) 
 3. [aisap Zig implementation](docs/aisap-zig.3.md) (DOCS WIP) (IMPLEMENTATION NOT YET USABLE)

(there's also some very early C bindings, which will be implemented in Zig. I
will begin working on the docs as soon as I feel the C API is sufficiently
usable.)

The ultimate goal is to have as many AppImages in
[aisap's internal library](profiles/README.md) as possible, while smaller, less
known apps may request their own permssions per the developer. Running programs
sandboxed should mostly be seamless and feel native with the system

For additional information on the permission system, see
[here](permissions/README.md)

As it is currently, the main aisap implementation requires a `squashfuse`
binary to function. I have attempted to create Go squashfuse bindings with
essentially zero success, so it will likely remain that way for the forseeable
future. Luckily, I have started working on a Zig implementation of aisap, and
due to Zig's extremely easy C interop, I already have some pretty decent Zig
squashfuse bindings to use. Don't expect the Zig implementation to be done
super soon, but it should be completely self-contained once it is and I will
probably replace the main CLI tool with it.
