const std = @import("std");
const io = std.io;
const fs = std.fs;
const span = std.mem.span;
const expect = std.testing.expect;

const Md5 = std.crypto.hash.Md5;

// TODO: figure out how to add this package correctly
const squashfs = @import("squashfuse-zig/lib.zig");
pub const SquashFs = squashfs.SquashFs;

const fuse = @import("squashfuse-zig/src/fuse.zig");
const E = fuse.E;
const linux = std.os.linux;

// Struct for holding our FUSE info
const Squash = struct {
    image: SquashFs,
    file_tree: std.StringArrayHashMap(SquashFs.Inode.Walker.Entry),
};

pub fn mountImage(src: []const u8, offset: usize) !void {
    var allocator = std.heap.c_allocator;

    const args = &[_]u8{
        "-s",
    };

    var squash = Squash{
        .image = SquashFs.init(
            allocator,
            src,
            .{ .offset = offset },
        ),
        .file_tree = std.StringArrayHashMap(SquashFs.Inode.Walker.Entry).init(allocator),
    };

    try fuse.main(allocator, args.items, &fuse_ops, squash);
}

export const fuse_ops = fuse.Operations{
    .init = squash_init,
    .getattr = squash_getattr,
    .getxattr = squash_getxattr,
    .open = squash_open,
    .opendir = squash_opendir,
    .release = squash_release,
    .releasedir = squash_releasedir,
    .create = squash_create,
    .read = squash_read,
    .readdir = squash_readdir,
    .readlink = squash_readlink,
};

fn squash_init(nfo: *fuse.ConnectionInfo, conf: *fuse.Config) callconv(.C) ?*anyopaque {
    _ = nfo;
    _ = conf;

    return fuse.context().private_data;
}

fn squash_read(p: [*:0]const u8, b: [*]u8, len: usize, o: std.os.linux.off_t, fi: *fuse.FileInfo) callconv(.C) c_int {
    _ = fi;

    var buf = b[0..len];

    const path = std.mem.span(p);
    var squash = fuse.privateDataAs(Squash);
    const offset: usize = @intCast(o);

    var entry = squash.file_tree.get(path[0..]) orelse return @intFromEnum(E.no_entry);
    var inode = entry.inode();
    inode.seekTo(offset) catch return @intFromEnum(E.io);

    const read_bytes = inode.read(buf) catch return @intFromEnum(E.io);

    return @intCast(read_bytes);
}

fn squash_create(_: [*:0]const u8, _: std.os.linux.mode_t, _: *fuse.FileInfo) callconv(.C) E {
    return .read_only;
}

fn squash_opendir(p: [*:0]const u8, fi: *fuse.FileInfo) callconv(.C) E {
    const path = std.mem.span(p);
    var squash = fuse.privateDataAs(Squash);

    if (std.mem.eql(u8, path, "/")) {
        var inode = squash.image.getRootInode();

        fi.handle = @intFromPtr(&inode.internal);

        return .success;
    }

    var entry = squash.file_tree.get(path[0..]) orelse return .no_entry;
    var inode = entry.inode();

    if (entry.kind != .directory) return .not_dir;

    fi.handle = @intFromPtr(&inode.internal);

    return .success;
}

fn squash_release(_: [*:0]const u8, fi: *fuse.FileInfo) callconv(.C) E {
    fi.handle = 0;
    return .success;
}

const squash_releasedir = squash_release;

fn squash_readdir(p: [*:0]const u8, filler: fuse.FillDir, _: linux.off_t, fi: *fuse.FileInfo, flags: fuse.ReadDirFlags) callconv(.C) E {
    _ = flags;
    _ = fi;

    var squash = fuse.privateDataAs(Squash);
    var path = std.mem.span(p);

    var root_inode = squash.image.getRootInode();
    var root_st = root_inode.statC() catch return .io;

    // Populate the current and parent directories
    filler.add(".", &root_st) catch return .io;
    filler.add("..", null) catch return .io;

    // Skip ahead to where the parent dir is in the hashmap
    var dir_idx: usize = undefined;
    if (std.mem.eql(u8, path, "/")) {
        dir_idx = 0;
    } else {
        dir_idx = squash.file_tree.getIndex(path) orelse return .no_entry;
    }

    const keys = squash.file_tree.keys();

    for (keys[dir_idx..]) |key| {
        const dirname = std.fs.path.dirname(key) orelse continue;

        if (key.len <= path.len) continue;
        if (!std.mem.eql(u8, path, key[0..path.len])) break;

        if (std.mem.eql(u8, path, dirname)) {
            var entry = squash.file_tree.get(key) orelse return .no_entry;
            var inode = squash.image.getInode(entry.id) catch return .io;

            // Load file info into buffer
            var st = inode.statC() catch return .io;

            var skip_slash: usize = 0;
            if (path.len > 1) skip_slash = 1;

            // This cast is normally not safe, but I've explicitly added a null
            // byte after the key slices upon creation
            const path_terminated: [*:0]const u8 = @ptrCast(key[dirname.len + skip_slash ..].ptr);

            try filler.add(path_terminated, &st);
        }
    }

    return .success;
}

fn squash_readlink(p: [*:0]const u8, b: [*]u8, len: usize) callconv(.C) E {
    var buf = b[0..len];

    const path = std.mem.span(p);
    var squash = fuse.privateDataAs(Squash);

    var entry = squash.file_tree.get(path) orelse return .no_entry;
    var inode = entry.inode();

    if (entry.kind != .sym_link) return .invalid_argument;

    inode.readLink(buf) catch return .io;

    return .success;
}

fn squash_open(p: [*:0]const u8, fi: *fuse.FileInfo) callconv(.C) E {
    const path = std.mem.span(p);
    var squash = fuse.privateDataAs(Squash);

    const entry = squash.file_tree.get(path) orelse return .no_entry;

    if (entry.kind == .directory) return .is_dir;

    fi.handle = @intFromPtr(&entry.inode().internal);
    fi.keep_cache = true;

    return .success;
}

// TODO
fn squash_getxattr(p: [*:0]const u8, r: [*:0]const u8, b: [*]u8, len: usize) callconv(.C) E {
    var buf = b[0..len];

    _ = p;
    _ = r;
    _ = buf;

    return .success;
}

fn squash_getattr(p: [*:0]const u8, stbuf: *std.os.linux.Stat, _: *fuse.FileInfo) callconv(.C) E {
    const path = std.mem.span(p);
    var squash = fuse.privateDataAs(Squash);

    // Load from the root inode
    if (std.mem.eql(u8, path, "/")) {
        var inode = squash.image.getRootInode();
        stbuf.* = inode.statC() catch return .io;

        return .success;
    }

    // Otherwise, grab the entry from our filetree hashmap
    var entry = squash.file_tree.get(path) orelse return .no_entry;
    var inode = entry.inode();

    stbuf.* = inode.statC() catch return .io;

    return .success;
}
