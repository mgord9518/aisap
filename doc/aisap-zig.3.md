# WIP DOCS -- THIS FILE NOT YET READY

# NAME

**aisap** - Zig package for sandboxing AppImage application bundles through bwrap

# SYNOPSIS

Adding as a dependency:

```zig
// In build.zig.zon:


```

Opening AppImage bundles:

```zig
const AppImage = aisap.AppImage;

// Returns an AppImage struct from a source path
AppImage.init(
    allocator: std.mem.Allocator,
    path: []const u8,
) !AppImage
```

AppImage methods:

```c
// Unmounts (if applicable) and cleans up the AppImage object
deinit(ai: *AppImage) void

// Returns the byte offset of the application bundle's filesystem image
offset(ai: *const AppImage) !u64

// Returns a base-16 encoded MD5sum of the application bundle's path, formatted in `buf`
// `buf` must be at least 33 bytes in size
md5(ai: *const AppImage, buf: []u8) ![:0]const u8

// NOT YET IMPLEMENTED
// This will return all individual arguments needed to pass to bwrap in order to sandbox the AppImage
// This normally shouldn't be called on its own
wrapArgs(ai: *const AppImage, allocator: std.mem.Allocator) ![]const [:0]const u8

// Returns the active permissions of the AppImage
permissions(ai: *const AppImage, allocator: std.mem.Allocator) ! Permissions

// NOT YET IMPLEMENTED
// This method will automatically create a temporary file in $XDG_RUNTIME_DIR/aisap and mount the application bundle's filesystem image to it
mount(ai: *AppImage) !void
```

# DESCRIPTION
