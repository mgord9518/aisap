// Simple program for testing my squashfuse bindings and aisap implementation
// I'll probably turn this into a re-implementation of the main squashfuse
// binary program after I get everything working

const std = @import("std");
const aisap = @import("aisap");
const fuse = @import("fuse.zig");
const E = fuse.E;

const AppImage = aisap.AppImage;
const SquashFs = aisap.SquashFs;

// Struct for holding our FUSE info
const Squash = struct {
    image: SquashFs,
    file_tree: std.StringArrayHashMap(SquashFs.Walker.Entry),
};

const test_ai_path = "/home/mgord9518/Downloads/Powder_Toy-96.2-x86_64.AppImage";

// I'll add this automatically soon, but currently, the `-s` flag must be
// supplied as it only works single-threaded
pub export fn main(argc: c_int, argv: [*:null]const ?[*:0]const u8) c_int {
    var ai = AppImage.init(test_ai_path) catch |err| {
        std.debug.print("error: {!}\n", .{err});
        return 2;
    };
    defer ai.deinit();

    var allocator = std.heap.c_allocator;
    var file_tree = std.StringArrayHashMap(SquashFs.Walker.Entry).init(allocator);

    var squash = Squash{ .image = ai.image, .file_tree = file_tree };

    var walker = squash.image.walk("") catch return 2;

    // Iterate over the AppImage's internal SquashFS image
    // This is to test my squashfuse bindings
    while (walker.next() catch return 1) |entry| {
        // Copy paths as they're automatically cleaned up by squashfuse and we
        // actually want them to stick around
        var new_path = allocator.alloc(u8, entry.path.len + 1) catch return 5;

        // TODO: implement this
        // Start new path with slash as squashfuse doesn't
        //        new_path[0] = '/';
        //        std.mem.copy(u8, new_path[1..], entry.path);

        std.mem.copy(u8, new_path, entry.path);

        // Make sure to add a null byte, because the keys will be not-so-optimally
        // casted to [*:0]const u8 for use in FUSE
        new_path[new_path.len - 1] = '\x00';
        new_path.len -= 1;

        // Now add to the HashMap
        const e = squash.file_tree.get(new_path);
        if (e == null) {
            squash.file_tree.put(new_path, entry) catch return 3;
        }

        // Debug
        //        var inode = squash.image.getInode(squash.file_tree.get(new_path).?.id) catch return 2;
        //        std.debug.print(" >>> {s}, {o}\n", .{ new_path, inode.internal.base.mode });
    }

    fuse.main(argc, argv, &fuse_ops, squash) catch return 1;

    return 0;
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

fn squash_read(p: [*:0]const u8, b: [*]u8, sz: usize, o: std.os.linux.off_t, fi: *fuse.FileInfo) callconv(.C) c_int {
    _ = fi;

    const path = std.mem.span(p);
    var squash = fuse.privateDataAs(Squash);
    const offset = @intCast(usize, o);

    // Build slice from buffer length and pointer
    var buf = b[0..sz];

    // FUSE provides leading slash, skip it as squashfuse doesn't
    var entry = squash.file_tree.get(path[1..]) orelse return @enumToInt(E.IO);
    var inode = squash.image.getInode(entry.id) catch return @enumToInt(E.IO);

    const read_bytes = squash.image.readRange(&inode, buf, offset) catch return @enumToInt(E.IO);

    return @intCast(c_int, read_bytes.len);
}

fn squash_create(p: [*:0]const u8, mode: std.os.linux.mode_t, fi: *fuse.FileInfo) callconv(.C) E {
    _ = p;
    _ = mode;
    _ = fi;

    return .ROFS;
}

fn squash_opendir(p: [*:0]const u8, fi: *fuse.FileInfo) callconv(.C) E {
    const path = std.mem.span(p);
    var squash = fuse.privateDataAs(Squash);

    if (std.mem.eql(u8, path, "/")) {
        var inode = squash.image.getInode(squash.image.internal.sb.root_inode) catch return .NOENT;

        fi.handle = @ptrToInt(&inode.internal);

        return .SUCCESS;
    }

    var entry = squash.file_tree.get(path[1..]) orelse return .NOENT;
    var inode = squash.image.getInode(entry.id) catch return .IO;

    if (entry.kind != .Directory) return .NOTDIR;

    fi.handle = @ptrToInt(&inode.internal);

    return .SUCCESS;
}

fn squash_release(p: [*:0]const u8, fi: *fuse.FileInfo) callconv(.C) E {
    _ = p;

    fi.handle = 0;
    return .SUCCESS;
}

const squash_releasedir = squash_release;

// TODO: refactor
fn squash_readdir(p: [*:0]const u8, buf: *anyopaque, filler: fuse.FillDir, offset: std.os.linux.off_t, fi: *fuse.FileInfo, flags: fuse.ReadDirFlags) callconv(.C) E {
    _ = flags;
    _ = offset;
    _ = fi;

    var squash = fuse.privateDataAs(Squash);

    // TODO: fix this to work without this janky shit
    var path = std.mem.span(p);

    var root_inode = squash.image.getInode(squash.image.internal.sb.root_inode) catch return .IO;
    var root_st = root_inode.statC() catch return .IO;
    //root_st.nlink = 2;

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

        if (it_depth <= path_depth and key.len > path.len and std.mem.eql(u8, key[0 .. path.len - 1], path[1..])) {
            var entry = squash.file_tree.get(key) orelse return .IO;
            var inode = squash.image.getInode(entry.id) catch return .IO;

            // Load file info into buffer
            var st = inode.statC() catch return .IO;

            var skip = path.len;
            if (path.len == 1) {
                skip = 0;
            }

            // This cast is normally not safe, but I've explicitly added a null
            // byte after the key slices upon creation
            _ = filler(buf, @ptrCast([*:0]const u8, key[skip..].ptr), &st, 0, 0);
        }
    }

    return .SUCCESS;
}

fn squash_readlink(p: [*:0]const u8, b: [*:0]u8, size: usize) callconv(.C) E {
    const path = std.mem.span(p);
    var squash = fuse.privateDataAs(Squash);

    var entry = squash.file_tree.get(path[1..]) orelse return .NOENT;
    var inode = squash.image.getInode(entry.id) catch return .NOENT;

    if (entry.kind != .SymLink) return .INVAL;

    var buf = b[0..size];

    inode.readLink(buf) catch return .IO;

    return .SUCCESS;
}

fn squash_init(nfo: *fuse.ConnectionInfo, conf: *fuse.Config) callconv(.C) ?*anyopaque {
    _ = nfo;
    _ = conf;

    return fuse.context().private_data;
}

fn squash_open(p: [*:0]const u8, fi: *fuse.FileInfo) callconv(.C) E {
    const path = std.mem.span(p);
    var squash = fuse.privateDataAs(Squash);

    const entry = squash.file_tree.get(path[1..]) orelse return .NOENT;
    var inode = squash.image.getInode(entry.id) catch return .NOENT;

    if (entry.kind == .Directory) return .ISDIR;

    fi.handle = @ptrToInt(&inode.internal);
    fi.keep_cache = 1;

    return .SUCCESS;
}

fn squash_getxattr(p: [*:0]const u8, r: [*:0]const u8, buf: [*:0]u8, i: usize) callconv(.C) E {
    _ = p;
    _ = r;
    _ = i;
    _ = buf;

    return .SUCCESS;
}

fn squash_getattr(p: [*:0]const u8, stbuf: *std.os.linux.Stat, fi: *fuse.FileInfo) callconv(.C) E {
    _ = fi;

    const path = std.mem.span(p);
    var squash = fuse.privateDataAs(Squash);

    // Load from the root inode
    if (std.mem.eql(u8, path, "/")) {
        var inode = squash.image.getInode(squash.image.internal.sb.root_inode) catch return .IO;
        stbuf.* = inode.statC() catch return .IO;

        return .SUCCESS;
    }

    // Otherwise, grab the entry from our filetree hashmap
    var entry = squash.file_tree.get(path[1..]) orelse return .IO;
    var inode = squash.image.getInode(entry.id) catch return .IO;

    stbuf.* = inode.statC() catch return .IO;

    return .SUCCESS;
}
