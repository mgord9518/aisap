// Simple program for testing my squashfuse bindings and aisap implementation
// I'll probably turn this into a re-implementation of the main squashfuse
// binary program after I get everything working

const std = @import("std");
const aisap = @import("aisap");
const fuse = @import("fuse.zig");

//var file_tree: std.StringHashMap(Walker.Entry) = undefined;

const AppImage = aisap.AppImage;
const SquashFs = aisap.SquashFs;

// Struct for holding our FUSE info
const Squash = struct {
    image: SquashFs,
    file_tree: std.StringArrayHashMap(SquashFs.Walker.Entry) = undefined,
};

// Been using my AppImage build of Go to test it, obviously if anyone else
// wants to use this pre-alpha test you'll need to change the path
//const test_ai_path = "/home/mgord9518/Downloads/YABG-0.0.1-x86_64.AppImage";
const test_ai_path = "/home/mgord9518/Downloads/Powder_Toy-96.2-x86_64.AppImage";
//const test_ai_path = "/home/mgord9518/Downloads/aisap-0.7.13-alpha-x86_64.AppImage";
//const test_ai_path = "/home/mgord9518/.local/bin/go";

pub export fn main(argc: c_int, argv: [*:null]const ?[*:0]const u8) c_int {
    var ai = AppImage.init(test_ai_path) catch |err| {
        std.debug.print("error: {!}\n", .{err});
        return 2;
    };
    defer ai.deinit();

    //    var open_appimage = SquashFs.init(test_ai_path, 594264) catch return 1;
    //    std.debug.print("ver: {}\n", .{ai.image.version});

    var squash = Squash{ .image = ai.image };

    var allocator = std.heap.c_allocator;
    var file_tree = std.StringArrayHashMap(SquashFs.Walker.Entry).init(allocator);

    squash.file_tree = file_tree;

    var walker = squash.image.walk("") catch return 2;

    // Iterate over the AppImage's internal SquashFS image
    // This is to test my squashfuse bindings
    while (walker.next() catch return 1) |entry| {
        // Copy paths as they're automatically cleaned up by squashfuse and we
        // actually want them to stick around
        var new_path = allocator.alloc(u8, entry.path.len) catch return 5;

        // TODO: implement this
        // Start new path with slash as squashfuse doesn't
        //        new_path[0] = '/';
        //        std.mem.copy(u8, new_path[1..], entry.path);

        //        var inode = fuse_ctx.image.getInode(entry.id) catch return 0;
        std.mem.copy(u8, new_path, entry.path);
        const e = squash.file_tree.get(new_path[0 .. new_path.len - 1]);
        if (e == null) {
            squash.file_tree.put(new_path[0 .. new_path.len - 1], entry) catch return 3;
        }

        //        std.debug.print(" >>> {s}, {o}\n", .{ new_path, inode.base.mode });
    }

    fuse.main(argc, argv, &fuse_ops, squash) catch return 1;

    return 0;
}

export const fuse_ops = fuse.Operations{
    .init = squash_init,
    .getattr = squash_getattr,
    .getxattr = squash_getxattr,
    .open = squash_open,
    .create = squash_create,
    .read = squash_read,
    .readdir = squash_readdir,
    .readlink = squash_readlink,
};

fn squash_read(p: [*:0]const u8, b: [*]u8, sz: usize, o: std.os.linux.off_t, fi: ?*fuse.FileInfo) callconv(.C) c_int {
    _ = fi;

    const path = std.mem.span(p);
    var squash = fuse.privateDataAs(Squash);
    const offset = @intCast(usize, o);

    // Build slice from buffer length and pointer
    var buf: []u8 = undefined;
    buf.ptr = b;
    buf.len = sz;

    // FUSE provides leading slash, skip it as squashfuse doesn't
    var entry = squash.file_tree.get(path[1..]) orelse return -5;
    var inode = squash.image.getInode(entry.id) catch return -5;

    const read_bytes = squash.image.readRange(&inode, buf, offset) catch return -5;

    return @intCast(c_int, read_bytes.len);
}

fn squash_create(p: [*:0]const u8, mode: std.os.linux.mode_t, fi: *fuse.FileInfo) callconv(.C) c_int {
    _ = p;
    _ = mode;
    _ = fi;

    return -30;
}

// TODO: refactor and fix.
// Does NOT currently read file trees correctly
fn squash_readdir(p: [*:0]const u8, buf: *anyopaque, filler: fuse.FillDir, offset: std.os.linux.off_t, fi: ?*fuse.FileInfo, flags: fuse.ReadDirFlags) callconv(.C) c_int {
    _ = flags;
    _ = offset;
    _ = fi;

    var squash = fuse.privateDataAs(Squash);

    // TODO: fix this to work without this janky shit
    var path = std.mem.span(p);

    var root_inode = squash.image.getInode(squash.image.internal.sb.root_inode) catch return -5;
    var root_st = root_inode.statC() catch return -5;
    //    root_st.nlink = 2;

    // Populate the current and parent directories
    _ = filler(buf, ".", &root_st, 0, 0);
    _ = filler(buf, "..", null, 0, 0);

    const keys = squash.file_tree.keys();
    for (keys) |key| {
        // Get the path depths of both the path provided by FUSE, and of our
        // current iteration
        var path_depth: u32 = 0;
        var it_depth: u32 = 0;
        for (path) |char| {
            if (char == '/') path_depth += 1;
        }

        for (key) |char| {
            if (char == '/') it_depth += 1;
        }

        //        std.debug.print("{d}, {s}, {s}\n", .{ path.len, path, key });

        if (it_depth <= path_depth and key.len > path.len and std.mem.eql(u8, key[0 .. path.len - 1], path[1..])) {
            var entry = squash.file_tree.get(key) orelse return -5;
            var inode = squash.image.getInode(entry.id) catch return 0;

            // Load file info into buffer
            var st = inode.statC() catch return 0;

            var skip = path.len;
            if (path.len == 1) {
                skip = 0;
            }

            _ = filler(buf, @ptrCast([*:0]const u8, key[skip..].ptr), &st, 0, 0);
        }
    }

    return 0;
}

fn squash_readlink(p: [*:0]const u8, b: [*:0]u8, size: usize) callconv(.C) c_int {
    const path = std.mem.span(p);
    var squash = fuse.privateDataAs(Squash);

    var entry = squash.file_tree.get(path[1..]) orelse return 1;
    var inode = squash.image.getInode(entry.id) catch return 0;

    // If not link type
    if (inode.internal.base.mode & 0o120000 != 0o120000) {
        return -22;
    }

    var buf = b[0..size];

    inode.readLink(buf) catch return -5;

    return 0;
}

fn squash_init(nfo: *fuse.ConnectionInfo, conf: *fuse.Config) callconv(.C) ?*anyopaque {
    _ = nfo;
    _ = conf;

    return fuse.context().private_data;
}

fn squash_open(p: [*:0]const u8, fi: ?*fuse.FileInfo) callconv(.C) c_int {
    const path = std.mem.span(p);
    var squash = fuse.privateDataAs(Squash);
    _ = fi;

    _ = squash.file_tree.get(path[1..]) orelse return -2;

    return 0;
}

fn squash_getxattr(p: [*:0]const u8, r: [*:0]const u8, buf: [*:0]u8, i: usize) callconv(.C) c_int {
    _ = p;
    _ = r;
    _ = i;
    _ = buf;

    return 0;
}

fn squash_getattr(p: [*:0]const u8, stbuf: *std.os.linux.Stat, fi: ?*fuse.FileInfo) callconv(.C) c_int {
    _ = fi;

    const path = std.mem.span(p);
    var squash = fuse.privateDataAs(Squash);

    // Load from the root inode and dip if FUSE wants the root permissions
    if (std.mem.eql(u8, path, "/")) {
        std.debug.print("root: {s}\n", .{path});
        var inode = squash.image.getInode(squash.image.internal.sb.root_inode) catch return -5;
        stbuf.* = inode.statC() catch return 1;
        //stbuf.nlink = 2;

        return 0;
    }

    // Otherwise, iterate through our entry map and find the path it wants
    const keys = squash.file_tree.keys();
    for (keys) |key| {
        //        const long_enough = key.len >= path.len - 1;
        const paths_eql = path.len > 0 and std.mem.eql(u8, path[1..], key);

        if (paths_eql) {
            var entry = squash.file_tree.get(key) orelse return -5;
            var inode = squash.image.getInode(entry.id) catch return 0;

            //std.debug.print("{s} {s} {}, {o}\n", .{ key, path, paths_eql, inode.base.mode });

            // Load file info into buffer
            stbuf.* = inode.statC() catch return 0;

            return 0;
        }
    }

    return 0;
}
