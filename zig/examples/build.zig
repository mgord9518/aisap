const std = @import("std");
const squashfuse = @import("squashfuse-zig/build.zig");

pub fn build(b: *std.build.Builder) void {
    // Standard target options allows the person running `zig build` to choose
    // what target to build for. Here we do not override the defaults, which
    // means any target is allowed, and the default is native. Other options
    // for restricting supported target set are available.
    const target = b.standardTargetOptions(.{});

    // Standard release options allow the person running `zig build` to select
    // between Debug, ReleaseSafe, ReleaseFast, and ReleaseSmall.
    const optimize = b.standardOptimizeOption(.{});

    const exe = b.addExecutable(.{
        .name = "example",
        .root_source_file = .{ .path = "src/main.zig" },
        .target = target,
        .optimize = optimize,
    });

    const exe_options = b.addOptions();
    exe_options.addOption(bool, "enable_xz", true);
    exe_options.addOption(bool, "enable_zlib", true);
    exe_options.addOption(bool, "use_libdeflate", true);
    exe_options.addOption(bool, "enable_lzo", false);
    exe_options.addOption(bool, "enable_lz4", true);
    exe_options.addOption(bool, "enable_zstd", true);
    exe_options.addOption(bool, "use_zig_zstd", false);

    const squashfuse_mod = b.addModule("squashfuse", .{
        .source_file = .{ .path = "../squashfuse-zig/lib.zig" },
        .dependencies = &.{
            .{
                .name = "build_options",
                .module = exe_options.createModule(),
            },
        },
    });

    const known_folders_mod = b.addModule("known-folders", .{
        .source_file = .{ .path = "../known-folders/known-folders.zig" },
    });

    const aisap_mod = b.addModule("aisap", .{
        .source_file = .{ .path = "../lib.zig" },
        .dependencies = &.{
            // TODO: handle this in aisap
            .{
                .name = "squashfuse",
                .module = squashfuse_mod,
            },
            .{
                .name = "known-folders",
                .module = known_folders_mod,
            },
        },
    });

    //    exe.addModule("squashfuse", squashfuse_mod);
    exe.addModule("aisap", aisap_mod);

    exe.addIncludePath("../../include");
    exe.addLibraryPath(".");

    squashfuse.linkVendored(exe, .{
        .enable_lz4 = true,
        .enable_lzo = false,
        .enable_zlib = true,
        .enable_zstd = true,
        .enable_xz = true,

        .use_libdeflate = true,
        //        .use_system_fuse = true,

        .squashfuse_dir = "squashfuse-zig",
    });

    exe.linkSystemLibrary("fuse3");

    b.installArtifact(exe);

    const run_cmd = b.addRunArtifact(exe);
    run_cmd.step.dependOn(b.getInstallStep());
    if (b.args) |args| {
        run_cmd.addArgs(args);
    }

    const run_step = b.step("run", "Run the app");
    run_step.dependOn(&run_cmd.step);

    // TODO: add tests
    const exe_tests = b.addTest(.{
        .root_source_file = .{ .path = "src/main.zig" },
        .target = target,
        .optimize = optimize,
    });

    const test_step = b.step("test", "Run unit tests");
    test_step.dependOn(&exe_tests.step);
}
