const std = @import("std");
const aisap = @import("../aisap.zig");

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
        .name = "squashfuse",
        .root_source_file = .{ .path = "src/main.zig" },
        .target = target,
        .optimize = optimize,
    });

    const squashfuse_mod = b.addModule("squashfuse", .{ .source_file = .{ .path = "../squashfuse-zig/lib.zig" } });
    const aisap_mod = b.addModule("aisap", .{ .source_file = .{ .path = "../lib.zig" } });

    exe.addModule("squashfuse", squashfuse_mod);
    exe.addModule("aisap", aisap_mod);

    exe.addIncludePath("../squashfuse-zig/squashfuse");
    exe.addIncludePath("../..");
    exe.addLibraryPath(".");

    // TODO: automatically include these when importing the bindings
    exe.addCSourceFile("../squashfuse-zig/squashfuse/cache.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/decompress.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/dir.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/file.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/fs.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/nonstd-makedev.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/nonstd-pread.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/nonstd-stat.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/stack.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/stat.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/swap.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/table.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/traverse.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/util.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/xattr.c", &[_][]const u8{});

    exe.linkLibC();
    // TODO: Submodule the source of these and import their code
    //    exe.linkSystemLibrary("cap");
    //    exe.linkSystemLibrary("bwrap.x86_64");
    exe.linkSystemLibrary("zlib");
    exe.linkSystemLibrary("zstd");
    exe.linkSystemLibrary("lz4");
    exe.linkSystemLibrary("lzma");
    exe.linkSystemLibrary("fuse3");

    // TODO: figure out why Zig automatically switches to dynamic linking
    // Linking libs as object files is a temporary solution which may be used
    // for now
    //exe.addObjectFile("/usr/lib/x86_64-linux-gnu/libzstd.a");

    exe.install();

    const run_cmd = exe.run();
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
