// Minimal FUSE wrapper

const std = @import("std");
const os = std.os;
const linux = os.linux;

const c = @cImport({
    @cDefine("_FILE_OFFSET_BITS", "64"); // Required for FUSE
    @cDefine("FUSE_USE_VERSION", "30");
    @cInclude("fuse3/fuse.h");
});

pub const FuseError = error{
    InvalidArgument,
    NoMountPoint,
    SetupFailed,
    MountFailed,
    DaemonizeFailed,
    SignalHandlerFailed,
    FileSystemError,
    UnknownError,
};

pub fn FuseErrorFromInt(err: c_int) FuseError!void {
    return switch (err) {
        0 => {},
        1 => FuseError.InvalidArgument,
        2 => FuseError.NoMountPoint,
        3 => FuseError.SetupFailed,
        4 => FuseError.MountFailed,
        5 => FuseError.DaemonizeFailed,
        6 => FuseError.SignalHandlerFailed,
        7 => FuseError.FileSystemError,

        else => FuseError.UnknownError,
    };
}

extern fn fuse_main_real(argc: c_int, argv: [*:null]const ?[*:0]const u8, op: *const Operations, op_size: usize, private_data: *const anyopaque) c_int;
pub fn main(argc: c_int, argv: [*:null]const ?[*:0]const u8, op: *const Operations, private_data: anytype) FuseError!void {
    try FuseErrorFromInt(fuse_main_real(argc, argv, op, @sizeOf(Operations), @ptrCast(*const anyopaque, &private_data)));
}

pub inline fn context() *Context {
    return c.fuse_get_context();
}

// Convenience function to fetch FUSE private data without casting
pub inline fn privateDataAs(comptime T: type) T {
    return @ptrCast(*T, @alignCast(@alignOf(T), context().private_data)).*;
}

pub const ReadDirFlags = c.fuse_readdir_flags;
pub const FileInfo = c.fuse_file_info;
pub const ConnectionInfo = c.fuse_conn_info;
pub const Config = c.fuse_config;
pub const Context = c.fuse_context;
pub const FillDirFlags = c.fuse_fill_dir_flags;
pub const PollHandle = c.fuse_pollhandle;
pub const BufVec = c.fuse_bufvec;

pub const StatVfs = c.struct_statvfs;

pub const FillDir = *const fn (*anyopaque, [*:0]const u8, ?*const os.Stat, linux.off_t, FillDirFlags) callconv(.C) c_int;

pub const Operations = extern struct {
    getattr: ?*const fn ([*:0]const u8, *os.Stat, ?*FileInfo) callconv(.C) c_int = null,
    readlink: ?*const fn ([*:0]const u8, [*:0]u8, usize) callconv(.C) c_int = null,
    mknod: ?*const fn ([*:0]const u8, linux.mode_t, linux.dev_t) callconv(.C) c_int = null,
    mkdir: ?*const fn ([*:0]const u8, linux.mode_t) callconv(.C) c_int = null,
    unlink: ?*const fn ([*:0]const u8) callconv(.C) c_int = null,
    rmdir: ?*const fn ([*:0]const u8) callconv(.C) c_int = null,
    symlink: ?*const fn ([*:0]const u8, [*:0]const u8) callconv(.C) c_int = null,
    rename: ?*const fn ([*:0]const u8, [*:0]const u8, c_uint) callconv(.C) c_int = null,
    link: ?*const fn ([*:0]const u8, [*:0]const u8) callconv(.C) c_int = null,
    chmod: ?*const fn ([*:0]const u8, linux.mode_t, ?*FileInfo) callconv(.C) c_int = null,
    chown: ?*const fn ([*:0]const u8, linux.uid_t, linux.gid_t, ?*FileInfo) callconv(.C) c_int = null,
    truncate: ?*const fn ([*:0]const u8, linux.off_t, ?*FileInfo) callconv(.C) c_int = null,
    open: ?*const fn ([*:0]const u8, ?*FileInfo) callconv(.C) c_int = null,
    read: ?*const fn ([*:0]const u8, [*]u8, usize, linux.off_t, ?*FileInfo) callconv(.C) c_int = null,
    write: ?*const fn ([*:0]const u8, [*:0]const u8, usize, linux.off_t, ?*FileInfo) callconv(.C) c_int = null,
    statfs: ?*const fn ([*:0]const u8, *StatVfs) callconv(.C) c_int = null,
    flush: ?*const fn ([*:0]const u8, ?*FileInfo) callconv(.C) c_int = null,
    release: ?*const fn ([*:0]const u8, *FileInfo) callconv(.C) c_int = null,
    fsync: ?*const fn ([*:0]const u8, c_int, *FileInfo) callconv(.C) c_int = null,
    setxattr: ?*const fn ([*:0]const u8, [*:0]const u8, [*:0]const u8, usize, c_int) callconv(.C) c_int = null,
    getxattr: ?*const fn ([*:0]const u8, [*:0]const u8, [*:0]u8, usize) callconv(.C) c_int = null,
    listxattr: ?*const fn ([*:0]const u8, [*:0]u8, usize) callconv(.C) c_int = null,
    removexattr: ?*const fn ([*:0]const u8, [*:0]const u8) callconv(.C) c_int = null,
    opendir: ?*const fn ([*:0]const u8, ?*FileInfo) callconv(.C) c_int = null,
    readdir: ?*const fn ([*:0]const u8, *anyopaque, FillDir, linux.off_t, ?*FileInfo, ReadDirFlags) callconv(.C) c_int = null,
    releasedir: ?*const fn ([*:0]const u8, ?*FileInfo) callconv(.C) c_int = null,
    fsyncdir: ?*const fn ([*:0]const u8, c_int, ?*FileInfo) callconv(.C) c_int = null,
    access: ?*const fn ([*:0]const u8, c_int) callconv(.C) c_int = null,
    init: ?*const fn (*ConnectionInfo, *Config) callconv(.C) ?*anyopaque = null,
    destroy: ?*const fn (*anyopaque) callconv(.C) void = null,
    create: ?*const fn ([*:0]const u8, linux.mode_t, *FileInfo) callconv(.C) c_int = null,
    lock: ?*const fn ([*:0]const u8, ?*FileInfo, c_int, *linux.Flock) callconv(.C) c_int = null,
    utimens: ?*const fn ([*:0]const u8, *const [2]linux.timespec, ?*FileInfo) callconv(.C) c_int = null,
    bmap: ?*const fn ([*:0]const u8, usize, *u64) callconv(.C) c_int = null,
    ioctl: ?*const fn ([*:0]const u8, c_int, *anyopaque, ?*FileInfo, c_uint, *anyopaque) callconv(.C) c_int = null,
    poll: ?*const fn ([*:0]const u8, ?*FileInfo, *PollHandle, *c_uint) callconv(.C) c_int = null,
    write_buf: ?*const fn ([*:0]const u8, *BufVec, linux.off_t, ?*FileInfo) callconv(.C) c_int = null,
    read_buf: ?*const fn ([*:0]const u8, [*c][*c]BufVec, usize, linux.off_t, ?*FileInfo) callconv(.C) c_int = null,
    flock: ?*const fn ([*:0]const u8, ?*FileInfo, c_int) callconv(.C) c_int = null,
    fallocate: ?*const fn ([*:0]const u8, c_int, linux.off_t, linux.off_t, ?*FileInfo) callconv(.C) c_int = null,
    copy_file_range: ?*const fn ([*:0]const u8, ?*FileInfo, linux.off_t, [*:0]const u8, ?*FileInfo, linux.off_t, usize, c_int) callconv(.C) isize = null,
    lseek: ?*const fn ([*:0]const u8, linux.off_t, c_int, ?*FileInfo) callconv(.C) linux.off_t = null,
};
