// Simple program for testing my squashfuse bindings and aisap implementation
// I'll probably turn this into a re-implementation of the main squashfuse
// binary program after I get everything working

const std = @import("std");
const aisap = @import("aisap");
const fuse = @import("fuse.zig");

const c = @cImport({
    @cInclude("stat.h"); // squashfuse (not system) stat header
    @cInclude("squashfuse.h");
});

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
const test_ai_path = "/home/mgord9518/Downloads/aisap-0.7.13-alpha-x86_64.AppImage";
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
        std.mem.copy(u8, new_path, entry.path);
        fuse_ctx.file_tree.put(new_path, entry) catch return 3;
    }

    _ = fuse.main(argc, argv, &fuse_ops, @ptrCast(*anyopaque, &fuse_ctx));

    return 0;
}

export const fuse_ops = fuse.Operations{
    .init = squash_init,
    .getattr = squash_getattr,
    .open = squash_open,
    .readdir = squash_readdir,
};

fn squash_readdir(p: [*:0]const u8, buf: *anyopaque, filler: fuse.FillDir, offset: std.os.linux.off_t, fi: ?*fuse.FileInfo, flags: fuse.ReadDirFlags) callconv(.C) c_int {
    _ = flags;
    _ = offset;
    _ = fi;

    var squash = @ptrCast(*Squash, @alignCast(8, fuse.getContext().private_data));

    // TODO: fix this to work without this janky shit
    var p1 = std.mem.span(p);
    var path: []const u8 = "";
    if (p1.len > 1) path = p1[0..p1.len];
    //    const path = std.mem.span(p);

    // Populate the current and parent directories
    _ = filler(buf, ".", null, 0, 0);
    _ = filler(buf, "..", null, 0, 0);

    // Iterate over the AppImage's internal SquashFS image
    // This is to test my squashfuse bindings
    var it = squash.file_tree.keyIterator();
    while (it.next()) |entry| {
        // Get the path depths of both the path provided by FUSE, and of our
        // current iteration
        var path_depth: u32 = 0;
        var it_depth: u32 = 0;
        for (path) |char| {
            if (char == '/') path_depth += 1;
        }

        for (entry.*) |char| {
            if (char == '/') it_depth += 1;
        }

        if (it_depth == path_depth) {
            // No clue why, but that data gets corrupted when I use '[:0]const u8'
            // as the type of `entry.path` instead of `[]const u8`. Because of
            // this, it must be casted even though we *know* it has a null terminator
            _ = filler(buf, @ptrCast([*:0]const u8, entry.*[path.len..].ptr), null, 0, 0);
        }
    }

    return 0;
}

fn squash_init(nfo: *fuse.ConnectionInfo, conf: *fuse.Config) callconv(.C) ?*anyopaque {
    _ = nfo;
    _ = conf;

    return fuse.getContext().private_data;
}

fn squash_open(path: [*:0]const u8, fi: ?*fuse.FileInfo) callconv(.C) c_int {
    _ = path;
    _ = fi;
    return 0;
}

fn squash_getattr(p: [*:0]const u8, stbuf: *std.os.linux.Stat, fi: ?*fuse.FileInfo) callconv(.C) c_int {
    _ = fi;

    const path = std.mem.span(p);
    var squash = @ptrCast(*Squash, @alignCast(8, fuse.getContext().private_data));

    // Load from the root inode and dip if FUSE wants the root permissions
    if (std.mem.eql(u8, path, "/")) {
        //        std.debug.print("root: {s}\n", .{path});
        var inode = squash.image.getInode(squash.image.internal.sb.root_inode) catch return 0;
        squash.image.statC(&inode, stbuf) catch return 0;

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
