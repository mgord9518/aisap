#!/bin/sh

# Example setup on integrating sandboxed AppImages into your shell as normal
# commands. This assumes that you store all of your AppImages in
# `~/Applications`, if that doesn't apply, simply modify the `appimage_dir`
# variable
appimage_dir=~/Applications

# In order to use this, either paste the code directly into your `.bashrc`, or
# keep it as a seperate file and load it using `. [SCRIPT PATH]` inside your
# `.bashrc` file

# Get latest version of AppImage if multiple are available as AppImageUpdate
# keeps old copies so it's likely multiple may exist
AppImage_get_latest() {
	set -- $(find "$appimage_dir" -maxdepth 1 -name "$1-*.AppImage" -o -name "$1-*.shImg")
	shift "$(($# - 1))"
	if [ $? -ne 0 ]; then
		echo "failed to find AppImage!"
		exit 1
	fi
	printf '%s\n' "$1"
}

# Make an alias for the desired launch command and the actual AppImage name
# Using this will automatically sandbox any aliased commands
AppImage_make_alias() {
	alias "$1"="aisap $(AppImage_get_latest $2)"
}

# Give `aisap` its own alias for easy launching
alias aisap=$(AppImage_get_latest aisap)

# Now create an alias (first argument) to the AppImage's name (second argument)
# If you have an AppImage named `The_Powder_Toy-x86_64.AppImage`, its name
# is `The_Powder_Toy`
AppImage_make_alias appimageupdate AppImageUpdate
AppImage_make_alias tpt            The_Powder_Toy
AppImage_make_alias brave          Brave

# Your AppImages should now be easily runnably from the shell using their
# aliases
