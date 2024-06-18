const std = @import("std");
const aisap = @import("aisap/zig/build.zig");

pub fn build(b: *std.Build) void {
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
        .root_source_file = b.path("src/main.zig"),
        .target = target,
        .optimize = optimize,
    });

    // TODO: find if there's some kind of convention here and follow it if so
    //const aisap_module = aisap.module(b);
    //exe.addModule("aisap", aisap_module);

    const aisap_dep = b.dependency("aisap", .{
        .target = target,
        .optimize = optimize,

        //        // These options will be renamed in the future
        //        .@"enable-fuse" = true,
        //        .@"enable-zlib" = true,
        //        //.@"use-zig-zlib" = true,
        //        .@"use-libdeflate" = true,
        //        .@"enable-xz" = true,
        //        .@"enable-lzma" = true,
        //        .@"enable-lzo" = false,
        //        .@"enable-lz4" = true,
        //        .@"enable-zstd" = true,
    });

    //    exe.root_module.addAnonymousImport("aisap", .{
    //        .root_source_file = .{ .path = "../lib.zig" },
    //    });
    exe.root_module.addImport("aisap", aisap_dep.module("aisap"));

    exe.linkLibrary(aisap_dep.artifact("zstd"));
    exe.linkLibrary(aisap_dep.artifact("lz4"));
    exe.linkLibrary(aisap_dep.artifact("deflate"));
    exe.linkLibrary(aisap_dep.artifact("fuse"));

    exe.linkLibC();

    //aisap.link(exe, .{});

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
        .root_source_file = b.path("src/main.zig"),
        .target = target,
        .optimize = optimize,
    });

    const test_step = b.step("test", "Run unit tests");
    test_step.dependOn(&exe_tests.step);
}
