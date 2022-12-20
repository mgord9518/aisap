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
    //    exe.addPackagePath("aisap", "../build.zig");
    exe.addPackagePath("squashfuse", "../squashfuse-zig/src/main.zig");
    exe.setTarget(target);
    exe.setBuildMode(mode);
    exe.linkLibC();
    //    exe.addIncludePath("/usr/include");
    //    exe.addIncludePath("/usr/include/x86_64-linux-gnu");
    //    exe.addIncludePath("/usr/lib/gcc/x86_64-linux-gnu/11/include/");
    exe.addIncludePath("../..");
    //    exe.addIncludePath("../libsquash-zig/libsquash/include");

    //    exe.addIncludePath("/usr/include/fuse3");
    //exe.addCSourceFile("../squashfuse/hl.c", &[_][]const u8{"-D_FILE_OFFSET_BITS=64"});
    exe.addLibraryPath(".");
    exe.addLibraryPath("../libsquash-zig/libsquash/build");
    exe.linkSystemLibrary("bwrap.x86_64");
    exe.linkSystemLibrary("squashfuse");
    exe.linkSystemLibrary("zlib");
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
