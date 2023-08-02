const std = @import("std");

pub const squashfuse = @import("squashfuse-zig/build.zig");

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

// TODO: maybe this should just be named `link`?
pub fn linkVendored(exe: *std.Build.Step.Compile, opts: LinkOptions) void {
    const prefix = thisDir();

    exe.addIncludePath(.{ .path = prefix ++ "/../include" });

    squashfuse.linkVendored(exe, .{
        .enable_lz4 = opts.enable_lz4,
        .enable_lzo = opts.enable_lzo,
        .enable_zlib = opts.enable_zlib,
        .enable_zstd = opts.enable_zstd,
        .enable_xz = opts.enable_xz,

        .use_libdeflate = opts.use_libdeflate,
    });
}

/// Returns a build module for aisap with deps pre-wrapped
pub fn module(b: *std.Build) *std.Build.Module {
    const prefix = thisDir();

    const lib_options = b.addOptions();
    lib_options.addOption(bool, "enable_xz", true);
    lib_options.addOption(bool, "enable_zlib", true);
    lib_options.addOption(bool, "use_libdeflate", true);
    lib_options.addOption(bool, "enable_lzo", false);
    lib_options.addOption(bool, "enable_lz4", true);
    lib_options.addOption(bool, "enable_zstd", true);
    lib_options.addOption(bool, "use_zig_zstd", false);

    const squashfuse_module = b.addModule("squashfuse", .{
        .source_file = .{ .path = prefix ++ "/squashfuse-zig/lib.zig" },
        .dependencies = &.{
            .{
                .name = "build_options",
                .module = lib_options.createModule(),
            },
        },
    });

    const known_folders_module = b.addModule("known-folders", .{
        .source_file = .{ .path = prefix ++ "/known-folders/known-folders.zig" },
    });

    return b.addModule("aisap", .{
        .source_file = .{ .path = prefix ++ "/lib.zig" },
        .dependencies = &.{
            .{
                .name = "squashfuse",
                .module = squashfuse_module,
            },
            .{
                .name = "known-folders",
                .module = known_folders_module,
            },
        },
    });
}

pub inline fn thisDir() []const u8 {
    return comptime std.fs.path.dirname(@src().file) orelse unreachable;
}

// Although this function looks imperative, note that its job is to
// declaratively construct a build graph that will be executed by an external
// runner.
pub fn build(b: *std.Build) void {
    const prefix = thisDir();

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
        .root_source_file = .{ .path = prefix ++ "/lib/c_api.zig" },
        .target = target,
        .optimize = optimize,
    });

    lib.addIncludePath(.{ .path = prefix ++ "/../include" });

    squashfuse.linkVendored(lib, .{
        .enable_lz4 = true,
        .enable_lzo = true,
        .enable_zlib = true,
        .enable_zstd = true,
        .enable_xz = true,

        .use_libdeflate = true,
    });

    const known_folders_mod = b.addModule("known-folders", .{
        .source_file = .{ .path = "known-folders/known-folders.zig" },
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
    lib.addModule("known-folders", known_folders_mod);

    const pie = b.option(bool, "pie", "build as a PIE (position independent executable)") orelse true;
    lib.pie = pie;

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
