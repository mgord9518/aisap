const std = @import("std");
const posix = std.posix;

pub fn main() !void {
    const allocator = std.heap.page_allocator;
    const string = "/toplevel/start${XDG_BIN_HOME}end/\\${MY_VAR}/other dir";

    var it = try SyntaxIterator.init(allocator, string);
    while (it.next()) |token| {
        std.debug.print("{s}\n", .{token.string});
        allocator.free(token.string);
    }
}

// TODO: tests and make this more robust
// It's alright for now since the permissions are curated, but it needs to be
// improved regardless
pub const SyntaxIterator = struct {
    allocator: std.mem.Allocator,

    pos: usize,
    slice: []const u8,
    args: std.ArrayList([]const u8),

    in_quotes: bool = false,

    pub const Token = union(enum) {
        string: []const u8,
        separator: Separator,

        pub const Separator = enum {
            pipe,
            whitespace,
        };
    };

    pub fn init(allocator: std.mem.Allocator, slice: []const u8) !SyntaxIterator {
        return .{
            .allocator = allocator,
            .pos = 0,
            .slice = slice,
            .args = std.ArrayList([]const u8).init(allocator),
        };
    }

    pub fn deinit(it: SyntaxIterator) void {
        it.args.deinit();
    }

    /// Breaks an input string into low-level tokens (strings and separators)
    /// These should be iterated over to create commands, variables, etc
    /// All strings must be freed by the caller
    pub fn next(it: *SyntaxIterator) ?Token {
        const has_whitespace = it.advanceToNextWord();
        if (has_whitespace) {
            return .{ .separator = .whitespace };
        }

        var tok_buf = std.ArrayList(u8).init(it.allocator);
        var var_buf = std.ArrayList(u8).init(it.allocator);
        defer var_buf.deinit();

        var backslash_escape = false;
        var in_variable = false;
        var var_state: enum {
            none,
            normal,
            bracket,
        } = .none;
        _ = &in_variable;

        var word_pos: usize = 0;

        for (it.slice[it.pos..]) |b| {
            var byte = b;
            var should_add = false;

            switch (byte) {
                '\\' => {
                    if (backslash_escape) {
                        should_add = true;
                    }

                    backslash_escape = !backslash_escape;
                },
                '$' => {
                    if (backslash_escape) {
                        should_add = true;
                        backslash_escape = false;
                    } else {
                        var_state = .normal;
                    }
                },
                '{' => {
                    switch (var_state) {
                        .normal => var_state = .bracket,
                        // TODO error
                        .bracket => unreachable,
                        .none => should_add = true,
                    }
                },
                '}' => {
                    const was_in_var = (var_state == .bracket);

                    switch (var_state) {
                        .normal => unreachable,
                        // TODO error
                        .bracket => var_state = .none,
                        .none => should_add = true,
                    }

                    if (was_in_var) {
                        _ = dumpVariableString(var_buf.items, tok_buf.writer());
                        var_buf.shrinkAndFree(0);
                    }
                },
                else => {
                    if (backslash_escape) {
                        backslash_escape = false;

                        byte = switch (byte) {
                            'n' => '\n',
                            't' => '\t',
                            'r' => '\r',
                            else => byte,
                        };
                    }

                    should_add = true;
                },
            }

            if (should_add) {
                if (var_state != .none) {
                    var_buf.append(byte) catch unreachable;
                } else {
                    tok_buf.append(byte) catch unreachable;
                }
            }

            word_pos += 1;

            // Return the current word if at the end of the string
            if (it.pos + word_pos >= it.slice.len) {
                if (var_state != .none) {
                    _ = dumpVariableString(var_buf.items, tok_buf.writer());
                    var_buf.shrinkAndFree(0);
                }

                defer it.pos += word_pos;
                return .{ .string = tok_buf.toOwnedSlice() catch unreachable };
            }
        }

        return null;
    }

    // Coerces a variable of any type into a string
    // Returns null if variable doesn't exist
    fn dumpVariableString(name: []const u8, writer: anytype) bool {
        if (posix.getenv(name)) |var_value| {
            _ = writer.write(var_value) catch unreachable;
            return true;
        }

        return false;
    }

    // Returns true if any whitespace was skipped
    fn advanceToNextWord(it: *SyntaxIterator) bool {
        const start_pos = it.pos;

        for (it.slice[it.pos..]) |byte| {
            if (byte != ' ') return start_pos != it.pos;
            it.pos += 1;
        }

        return start_pos != it.pos;
    }
};
