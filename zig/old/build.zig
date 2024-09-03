const std = @import("std");

pub const LinkOptions = struct {
    enable_zstd: bool = true,
    enable_lz4: bool = true,
    enable_lzo: bool = false,
    enable_zlib: bool = true,
    enable_xz: bool = true,

    enable_fuse: bool = true,
    use_system_fuse: bool = true,

    use_libdeflate: bool = true,
};

// Although this function looks imperative, note that its job is to
// declaratively construct a build graph that will be executed by an external
// runner.
pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    const lib = b.addStaticLibrary(.{
        .name = "aisap",
        .root_source_file = b.path("lib/c_api.zig"),
        .target = target,
        .optimize = optimize,
    });

    lib.addIncludePath(b.path("../include"));

    const squashfuse_dep = b.dependency("squashfuse", .{
        .target = target,
        .optimize = optimize,

        .zlib_decompressor = .libdeflate_static,
        .xz_decompressor = .liblzma_static,
        .lz4_decompressor = .liblz4_static,
        .zstd_decompressor = .libzstd_static,
    });

    const fuse_dep = b.dependency("fuse", .{
        .target = target,
        .optimize = optimize,
    });

    const known_folders_dep = b.dependency("known_folders", .{
        .target = target,
        .optimize = optimize,
    });

    lib.root_module.addImport(
        "squashfuse",
        squashfuse_dep.module("squashfuse"),
    );

    lib.root_module.addImport(
        "known-folders",
        known_folders_dep.module("known-folders"),
    );

    _ = b.addModule("aisap", .{
        .root_source_file = b.path("lib.zig"),
        .imports = &.{
            .{
                .name = "squashfuse",
                .module = squashfuse_dep.module("squashfuse"),
            },
            .{
                .name = "fuse",
                .module = fuse_dep.module("fuse"),
            },
        },
    });

    lib.root_module.addImport(
        "fuse",
        fuse_dep.module("fuse"),
    );

    //    lib.root_module.addImport("known-folders", known_folders_module);

    //    const pie = b.option(bool, "pie", "build as a PIE (position independent executable)") orelse true;
    //    lib.pie = pie;

    lib.linkLibC();

    // This declares intent for the library to be installed into the standard
    // location when the user invokes the "install" step (the default step when
    // running `zig build`).
    b.installArtifact(lib);
    b.installArtifact(squashfuse_dep.artifact("deflate"));
    b.installArtifact(squashfuse_dep.artifact("zstd"));
    b.installArtifact(squashfuse_dep.artifact("lz4"));
    b.installArtifact(fuse_dep.artifact("fuse"));

    // Creates a step for unit testing. This only builds the test executable
    // but does not run it.
    const main_tests = b.addTest(.{
        .root_source_file = b.path("lib.zig"),
        .target = target,
        .optimize = optimize,
    });

    const run_main_tests = b.addRunArtifact(main_tests);

    // This creates a build step. It will be visible in the `zig build --help` menu,
    // and can be selected like this: `zig build test`
    // This will evaluate the `test` step rather than the default, which is "install".
    const test_step = b.step("test", "Run library tests");
    test_step.dependOn(&run_main_tests.step);
}
