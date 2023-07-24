const std = @import("std");
const squashfuse = @import("squashfuse-zig/build.zig");

// Although this function looks imperative, note that its job is to
// declaratively construct a build graph that will be executed by an external
// runner.
pub fn build(b: *std.Build) void {
    // Standard target options allows the person running `zig build` to choose
    // what target to build for. Here we do not override the defaults, which
    // means any target is allowed, and the default is native. Other options
    // for restricting supported target set are available.
    const target = b.standardTargetOptions(.{});

    // Standard optimization options allow the person running `zig build` to select
    // between Debug, ReleaseSafe, ReleaseFast, and ReleaseSmall. Here we do not
    // set a preferred release mode, allowing the user to decide how to optimize.
    const optimize = b.standardOptimizeOption(.{});

    const lib_options = b.addOptions();
    lib_options.addOption(bool, "enable_xz", true);
    lib_options.addOption(bool, "enable_zlib", true);
    lib_options.addOption(bool, "use_libdeflate", true);
    lib_options.addOption(bool, "enable_lzo", false);
    lib_options.addOption(bool, "enable_lz4", true);
    lib_options.addOption(bool, "enable_zstd", true);
    lib_options.addOption(bool, "use_zig_zstd", false);

    const lib = b.addStaticLibrary(.{
        .name = "aisap",
        // In this case the main source file is merely a path, however, in more
        // complicated build scripts, this could be a generated file.
        .root_source_file = .{ .path = "lib/c_api.zig" },
        .target = target,
        .optimize = optimize,
    });

    const pie = b.option(bool, "no-pie", "build as a PIE (position independent executable)") orelse true;
    lib.pie = !pie;

    squashfuse.linkVendored(lib, .{
        .enable_lz4 = true,
        .enable_lzo = true,
        .enable_zlib = true,
        .enable_zstd = true,
        .enable_xz = true,

        .use_libdeflate = true,

        .squashfuse_dir = "squashfuse-zig",
    });

    const squashfuse_mod = b.addModule("squashfuse", .{
        .source_file = .{ .path = "squashfuse-zig/lib.zig" },
        .dependencies = &.{
            .{
                .name = "build_options",
                .module = lib_options.createModule(),
            },
        },
    });

    lib.addModule("squashfuse", squashfuse_mod);

    lib.addIncludePath("..");

    lib.linkLibC();

    // This declares intent for the library to be installed into the standard
    // location when the user invokes the "install" step (the default step when
    // running `zig build`).
    b.installArtifact(lib);

    // Creates a step for unit testing. This only builds the test executable
    // but does not run it.
    const main_tests = b.addTest(.{
        .root_source_file = .{ .path = "lib.zig" },
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
