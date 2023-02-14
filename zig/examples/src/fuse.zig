// Basic FUSE wrapper
// TODO: Re-wrap a lot of this stuff to make it more idiomatic as it's
// currently very C-esque

const std = @import("std");
const warn = @import("std").debug.print;
const aisap = @import("aisap");

const os = std.os;
const mode_t = std.os.linux.mode_t;
const size_t = usize;
const ssize_t = isize;
const off_t = std.os.linux.off_t;
const dev_t = std.os.linux.dev_t;
const uid_t = std.os.linux.uid_t;
const gid_t = std.os.linux.gid_t;
const timespec = std.os.linux.timespec;
const ino_t = std.os.linux.ino_t;

// TODO: declare fuse types to pull into Zig projects
const c = @cImport({
    @cDefine("_FILE_OFFSET_BITS", "64");
    @cInclude("stat.h"); // squashfuse (not system) stat header
    @cInclude("squashfuse.h");
    @cInclude("fuse3/fuse.h");
});

pub export fn run(argc: c_int, argv: [*:null]const ?[*:0]const u8, op: *const Operations, private_data: ?*anyopaque) c_int {
    return fuse_main_real(argc, argv, op, @sizeOf(Operations), private_data);
}

extern fn fuse_main_real(argc: c_int, argv: [*:null]const ?[*:0]const u8, op: *const Operations, op_size: size_t, private_data: ?*anyopaque) c_int;

pub const ReaddirFlags = c.fuse_readdir_flags;

pub const Operations = extern struct {
    getattr: ?*const fn ([*:0]const u8, *os.Stat, *FileInfo) callconv(.C) c_int = null,
    readlink: ?*const fn ([*:0]const u8, [*:0]u8, size_t) callconv(.C) c_int = null,
    mknod: ?*const fn ([*:0]const u8, mode_t, dev_t) callconv(.C) c_int = null,
    mkdir: ?*const fn ([*:0]const u8, mode_t) callconv(.C) c_int = null,
    unlink: ?*const fn ([*:0]const u8) callconv(.C) c_int = null,
    rmdir: ?*const fn ([*:0]const u8) callconv(.C) c_int = null,
    symlink: ?*const fn ([*:0]const u8, [*:0]const u8) callconv(.C) c_int = null,
    rename: ?*const fn ([*:0]const u8, [*:0]const u8, c_uint) callconv(.C) c_int = null,
    link: ?*const fn ([*:0]const u8, [*:0]const u8) callconv(.C) c_int = null,
    chmod: ?*const fn ([*:0]const u8, mode_t, *FileInfo) callconv(.C) c_int = null,
    chown: ?*const fn ([*:0]const u8, uid_t, gid_t, *FileInfo) callconv(.C) c_int = null,
    truncate: ?*const fn ([*:0]const u8, off_t, ?*FileInfo) callconv(.C) c_int = null,
    open: ?*const fn ([*:0]const u8, *FileInfo) callconv(.C) c_int = null,
    read: ?*const fn ([*:0]const u8, [*:0]u8, size_t, off_t, *FileInfo) callconv(.C) c_int = null,
    write: ?*const fn ([*:0]const u8, [*:0]const u8, size_t, off_t, *FileInfo) callconv(.C) c_int = null,
    statfs: ?*const fn ([*:0]const u8, *extrn.Statvfs) callconv(.C) c_int = null,
    flush: ?*const fn ([*:0]const u8, ?*FileInfo) callconv(.C) c_int = null,
    release: ?*const fn ([*:0]const u8, *FileInfo) callconv(.C) c_int = null,
    fsync: ?*const fn ([*:0]const u8, c_int, *FileInfo) callconv(.C) c_int = null,
    setxattr: ?*const fn ([*:0]const u8, [*:0]const u8, [*:0]const u8, size_t, c_int) callconv(.C) c_int = null,
    getxattr: ?*const fn ([*:0]const u8, [*:0]const u8, [*:0]u8, size_t) callconv(.C) c_int = null,
    listxattr: ?*const fn ([*:0]const u8, [*:0]u8, size_t) callconv(.C) c_int = null,
    removexattr: ?*const fn ([*:0]const u8, [*:0]const u8) callconv(.C) c_int = null,
    opendir: ?*const fn ([*:0]const u8, ?*FileInfo) callconv(.C) c_int = null,
    readdir: ?*const fn ([*:0]const u8, *anyopaque, FillDir, off_t, *FileInfo, ReaddirFlags) callconv(.C) c_int = null,
    releasedir: ?*const fn ([*:0]const u8, ?*FileInfo) callconv(.C) c_int = null,
    fsyncdir: ?*const fn ([*:0]const u8, c_int, ?*FileInfo) callconv(.C) c_int = null,
    access: ?*const fn ([*:0]const u8, c_int) callconv(.C) c_int = null,
    init: ?*const fn (*c.fuse_conn_info, *c.fuse_config) callconv(.C) ?*anyopaque = null,
    destroy: ?*const fn (*anyopaque) callconv(.C) void = null,
    create: ?*const fn ([*:0]const u8, mode_t, *FileInfo) callconv(.C) c_int = null,
    lock: ?*const fn ([*:0]const u8, ?*FileInfo, c_int, *std.os.linux.Flock) callconv(.C) c_int = null,
    utimens: ?*const fn ([*:0]const u8, *const [2]timespec, ?*FileInfo) callconv(.C) c_int = null,
    bmap: ?*const fn ([*:0]const u8, size_t, *u64) callconv(.C) c_int = null,
    ioctl: ?*const fn ([*:0]const u8, c_int, *anyopaque, ?*FileInfo, c_uint, *anyopaque) callconv(.C) c_int = null,
    poll: ?*const fn ([*:0]const u8, ?*FileInfo, *c.fuse_pollhandle, *c_uint) callconv(.C) c_int = null,
    write_buf: ?*const fn ([*:0]const u8, *c.fuse_bufvec, off_t, ?*FileInfo) callconv(.C) c_int = null,
    read_buf: ?*const fn ([*:0]const u8, [*c][*c]c.fuse_bufvec, size_t, off_t, ?*FileInfo) callconv(.C) c_int = null,
    flock: ?*const fn ([*:0]const u8, ?*FileInfo, c_int) callconv(.C) c_int = null,
    fallocate: ?*const fn ([*:0]const u8, c_int, off_t, off_t, ?*FileInfo) callconv(.C) c_int = null,
    copy_file_range: ?*const fn ([*:0]const u8, ?*FileInfo, off_t, [*:0]const u8, ?*FileInfo, off_t, size_t, c_int) callconv(.C) ssize_t = null,
    lseek: ?*const fn ([*:0]const u8, off_t, c_int, ?*FileInfo) callconv(.C) off_t = null,
};

pub const FileInfo = extern struct {
    flags: c_int,
    bitfield0: packed struct(c_uint) {
        writepage: u1,
        direct_io: u1,
        keep_cache: u1,
        flush: u1,
        nonseekable: u1,
        flock_release: u1,
        cache_readdir: u1,
        noflush: u1,
        padding: u24,
    },
    bitfield1: packed struct(c_uint) {
        padding2: u32,
    },
    fh: u64,
    lock_owner: u64,
    poll_events: u32,
};

extern threadlocal var errno: c_int;
const extrn = struct {
    const Statvfs = extern struct {
        f_bsize: c_ulong,
        f_frsize: c_ulong,
        f_blocks: fsblkcnt_t,
        f_bfree: fsblkcnt_t,
        f_bavail: fsblkcnt_t,
        f_files: fsfilcnt_t,
        f_ffree: fsfilcnt_t,
        f_favail: fsfilcnt_t,
        f_fsid: c_ulong,
        f_flag: c_ulong,
        f_namemax: c_ulong,
    };
    const fsblkcnt_t = c_ulonglong;
    const fsfilcnt_t = c_ulonglong;
    comptime {
        std.debug.assert(@sizeOf(usize) == @sizeOf(u64)); // only 64bit host is currently supported
        // on 32bit fsblkcnt_t/fsfilcnt_t are c_ulong
    }
    const DIR = opaque {};
    extern fn opendir(name: [*:0]const u8) ?*DIR;
    extern fn readdir(dirp: *DIR) ?*dirent;
    extern fn closedir(dirp: *DIR) c_int;
    extern fn mkfifoat(dirfd: c_int, pathname: [*:0]const u8, mode: mode_t) c_int;
    extern fn lstat(pathname: [*:0]const u8, statbuf: *os.Stat) c_int;
    extern fn access(pathname: [*:0]const u8, mode: c_int) c_int;
    extern fn readlink(path: [*:0]const u8, buf: [*:0]u8, bufsiz: size_t) ssize_t;
    extern fn openat(dirfd: c_int, pathname: [*:0]const u8, flags: c_int, mode: mode_t) c_int;
    extern fn close(fd: c_int) c_int;
    extern fn mkdirat(dirfd: c_int, pathname: [*:0]const u8, mode: mode_t) c_int;
    extern fn symlinkat(oldpath: ?[*:0]const u8, newdirfd: c_int, newpath: [*:0]const u8) c_int;
    extern fn mknodat(dirfd: c_int, pathname: [*:0]const u8, mode: mode_t, dev: dev_t) c_int;
    extern fn mkdir(path: [*:0]const u8, mode: mode_t) c_int;
    extern fn link(oldpath: [*:0]const u8, newpath: [*:0]const u8) c_int;
    extern fn chmod(pathname: [*:0]const u8, mode: mode_t) c_int;
    extern fn lchown(pathname: [*:0]const u8, owner: uid_t, group: gid_t) c_int;
    extern fn ftruncate(fd: c_int, length: off_t) c_int;
    extern fn truncate(path: [*:0]const u8, length: off_t) c_int;
    extern fn open(pathname: [*:0]const u8, flags: c_int, ...) c_int;
    extern fn pread(fd: c_int, buf: [*:0]u8, count: size_t, offset: off_t) ssize_t;
    extern fn pwrite(fd: c_int, buf: [*:0]const u8, count: size_t, offset: off_t) ssize_t;
    extern fn statvfs(path: [*:0]const u8, buf: *Statvfs) c_int;
    extern fn lseek(fd: c_int, offset: off_t, whence: c_int) off_t;

    const dirent = extern struct {
        d_ino: ino_t,
        d_off: off_t,
        d_reclen: c_ushort,
        d_type: u8,
        d_name: [256]u8,
    };
};

pub const FillDir = *const fn (*anyopaque, [*:0]const u8, ?*const c.struct_stat, off_t, c.fuse_fill_dir_flags) callconv(.C) c_int;
