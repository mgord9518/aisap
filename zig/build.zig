const std = @import("std");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    const squashfuse_dep = b.dependency("squashfuse", .{
        .target = target,
        .optimize = optimize,

        .zlib_decompressor = .libdeflate_static,
        .zstd_decompressor = .libzstd_static,
        .lz4_decompressor = .liblz4_static,
        .xz_decompressor = .liblzma_static,
    });

    const aisap_module = b.addModule("aisap", .{
        .root_source_file = b.path("lib/AppImage.zig"),
        .imports = &.{
            .{
                .name = "squashfuse",
                .module = squashfuse_dep.module("squashfuse"),
            },
        },
    });

    const exe = b.addExecutable(.{
        .name = "aisap",
        .root_source_file = b.path("src/main.zig"),
        .target = target,
        .optimize = optimize,
    });

    exe.linkLibC();

    exe.root_module.addImport(
        "aisap",
        aisap_module,
    );

    exe.linkLibrary(squashfuse_dep.artifact("zstd"));
    exe.linkLibrary(squashfuse_dep.artifact("deflate"));
    exe.linkLibrary(squashfuse_dep.artifact("lzma"));
    exe.linkLibrary(squashfuse_dep.artifact("lz4"));

    b.installArtifact(exe);
}
