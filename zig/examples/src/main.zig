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
    file_tree: std.StringHashMap(SquashFs.Walker.Entry) = undefined,
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
    std.debug.print("ver: {}\n", .{ai.image.version});

    var fuse_ctx = Squash{ .image = ai.image };

    var allocator = std.heap.c_allocator;
    var file_tree = std.StringHashMap(SquashFs.Walker.Entry).init(allocator);

    fuse_ctx.file_tree = file_tree;

    var walker = fuse_ctx.image.walk("") catch return 2;

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

        // A bit hacky, this copies over the whole string (including null
        // terminator) but only sets the slice to use the bytes leading up to
        // the terminator, not including it. This allows for easy byte
        // comparisons in Zig, but also allows casting to [*:0]const u8 for use
        // in C code.
        std.mem.copy(u8, new_path, entry.path);
        fuse_ctx.file_tree.put(new_path[0 .. new_path.len - 1], entry) catch return 3;

        std.debug.print(" >>> {s}\n", .{new_path});
    }

    fuse.main(argc, argv, &fuse_ops, fuse_ctx) catch return 1;

    return 0;
}

export const fuse_ops = fuse.Operations{
    .init = squash_init,
    .getattr = squash_getattr,
    .open = squash_open,
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
    var entry = squash.file_tree.get(path[1..]) orelse return 1;
    var inode = squash.image.getInode(entry.id) catch return 0;

    var written: usize = undefined;
    written = squash.image.readRange(&inode, buf, offset) catch return -1;

    return @intCast(c_int, written);
}

// TODO: refactor and fix.
// Does NOT currently read file trees correctly
fn squash_readdir(p: [*:0]const u8, buf: *anyopaque, filler: fuse.FillDir, offset: std.os.linux.off_t, fi: ?*fuse.FileInfo, flags: fuse.ReadDirFlags) callconv(.C) c_int {
    _ = flags;
    _ = offset;
    _ = fi;

    var squash = fuse.privateDataAs(Squash);

    // TODO: fix this to work without this janky shit
    var p1 = std.mem.span(p);
    var path: []const u8 = "";
    if (p1.len > 1) path = p1[0..p1.len];
    //    const path = std.mem.span(p);

    // Populate the current and parent directories
    _ = filler(buf, ".", null, 0, 0);
    _ = filler(buf, "..", null, 0, 0);

    var st = std.mem.zeroes(std.os.Stat);

    // Iterate over the AppImage's internal SquashFS image
    // This is to test my squashfuse bindings
    var it = squash.file_tree.keyIterator();
    while (it.next()) |val| {
        // Get the path depths of both the path provided by FUSE, and of our
        // current iteration
        var path_depth: u32 = 0;
        var it_depth: u32 = 0;
        for (path) |char| {
            if (char == '/') path_depth += 1;
        }

        for (val.*) |char| {
            if (char == '/') it_depth += 1;
        }

        if (it_depth == path_depth) {
            var entry = squash.file_tree.get(val.*) orelse return 1;
            var inode = squash.image.getInode(entry.id) catch return 0;

            // Load file info into buffer
            squash.image.statC(&inode, &st) catch return 0;

            // No clue why, but that data gets corrupted when I use '[:0]const u8'
            // as the type of `entry.path` instead of `[]const u8`. Because of
            // this, it must be casted even though we *know* it has a null terminator
            _ = filler(buf, @ptrCast([*:0]const u8, val.*[path.len..].ptr), &st, 0, 0);
        }
    }

    return 0;
}

fn squash_readlink(p: [*:0]const u8, buf: [*:0]u8, size: usize) callconv(.C) c_int {
    const path = std.mem.span(p);
    var squash = fuse.privateDataAs(Squash);
    _ = size;

    var entry = squash.file_tree.get(path[1..]) orelse return 1;
    var inode = squash.image.getInode(entry.id) catch return 0;

    // If not link type
    if (inode.base.mode & 0o120000 != 0o120000) {
        return 1;
    }

    var sz: usize = undefined;
    _ = squash.image.readlink(&inode, buf, &sz);

    return 0;
}

fn squash_init(nfo: *fuse.ConnectionInfo, conf: *fuse.Config) callconv(.C) ?*anyopaque {
    _ = nfo;
    _ = conf;

    return fuse.context().private_data;
}

fn squash_open(path: [*:0]const u8, fi: ?*fuse.FileInfo) callconv(.C) c_int {
    _ = path;
    _ = fi;

    return 0;
}

fn squash_getattr(p: [*:0]const u8, stbuf: *std.os.linux.Stat, fi: ?*fuse.FileInfo) callconv(.C) c_int {
    _ = fi;

    const path = std.mem.span(p);
    var squash = fuse.privateDataAs(Squash);

    // Load from the root inode and dip if FUSE wants the root permissions
    if (std.mem.eql(u8, path, "/")) {
        //        std.debug.print("root: {s}\n", .{path});
        var inode = squash.image.getInode(squash.image.internal.sb.root_inode) catch return 0;
        squash.image.statC(&inode, stbuf) catch return 0;
        stbuf.nlink = 2;

        return 0;
    }

    // Otherwise, iterate through our entry map and find the path it wants
    var it = squash.file_tree.keyIterator();
    while (it.next()) |val| {
        const long_enough = val.*.len >= path.len - 1;
        const paths_eql = long_enough and std.mem.eql(u8, path[1..], val.*[0 .. path.len - 1]);

        if (paths_eql) {
            var entry = squash.file_tree.get(val.*) orelse return 1;
            var inode = squash.image.getInode(entry.id) catch return 0;

            // Load file info into buffer
            squash.image.statC(&inode, stbuf) catch return 0;
        }
    }

    return 0;
}
