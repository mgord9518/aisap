const std = @import("std");

pub const ProfileParser = struct {
    var variable_map = std.StringHashMap([]const u8);

    pub fn init(allocator: std.mem.Allocator) !ProfileParser {
        return .{
            .variable_map = try std.StringHashMap([]const u8).init(allocator),
        };
    }
};
