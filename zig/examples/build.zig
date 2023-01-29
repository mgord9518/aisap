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
    const mode = b.standardReleaseOptions();

    const exe = b.addExecutable("open", "src/main.zig");
    exe.addPackagePath("aisap", "../lib.zig");
    exe.addPackagePath("squashfuse", "../squashfuse-zig/src/main.zig");
    exe.setTarget(target);
    exe.setBuildMode(mode);
    exe.linkLibC();

    exe.addIncludePath("../squashfuse-zig/squashfuse");
    exe.addIncludePath("../..");
    exe.addLibraryPath(".");

    // TODO: automatically include these when importing the bindings
    exe.addCSourceFile("../squashfuse-zig/squashfuse/stat.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/nonstd-makedev.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/nonstd-pread.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/nonstd-stat.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/swap.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/fs.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/file.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/util.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/traverse.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/stack.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/dir.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/decompress.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/cache.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/table.c", &[_][]const u8{});
    exe.addCSourceFile("../squashfuse-zig/squashfuse/xattr.c", &[_][]const u8{});

    exe.linkSystemLibrary("bwrap.x86_64");
    exe.linkSystemLibrary("zlib");
    exe.linkSystemLibrary("zstd");
    exe.linkSystemLibrary("lz4");
    exe.linkSystemLibrary("lzma");

    exe.linkSystemLibrary("cap");
    exe.install();

    const run_cmd = exe.run();
    run_cmd.step.dependOn(b.getInstallStep());
    if (b.args) |args| {
        run_cmd.addArgs(args);
    }

    const run_step = b.step("run", "Run the app");
    run_step.dependOn(&run_cmd.step);

    const exe_tests = b.addTest("src/main.zig");
    exe_tests.setTarget(target);
    exe_tests.setBuildMode(mode);

    const test_step = b.step("test", "Run unit tests");
    test_step.dependOn(&exe_tests.step);
}
