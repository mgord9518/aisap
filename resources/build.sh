#!/bin/sh
# Quick and dirty shell script to build aisap AppDir from source

[ -z "$ARCH" ] && ARCH=$(uname -m)
PATH="$PATH:$HOME/.local/bin"
export VERSION=$(cat zig/build.zig.zon | grep ' .version' | cut -d'"' -f2)

if [ ! $(command -v 'go') ]; then
	echo 'Failed to locate GoLang compiler! Unable to build'
	exit 1
fi

# Get mkappimage
wget 'https://raw.githubusercontent.com/mgord9518/appimage_scripts/main/scripts/get_mkappimage.sh'
. ./get_mkappimage.sh

aisapUrl='github.com/mgord9518/aisap'
aisapRawUrl='raw.githubusercontent.com/mgord9518/aisap/main'

mkdir -p 'AppDir/usr/bin' \
         'AppDir/usr/share/metainfo' \
         'AppDir/usr/share/icons/hicolor/scalable/apps'

cd 'cmd/aisap'

# Use local files for building
echo 'replace github.com/mgord9518/aisap => ../../
replace github.com/mgord9518/aisap/permissions => ../../permissions
replace github.com/mgord9518/aisap/profiles => ../../profiles
replace github.com/mgord9518/aisap/spooky => ../../spooky
replace github.com/mgord9518/aisap/helpers => ../../helpers
' >> go.mod
go mod tidy

CGO_ENABLED=0 go build \
    -o '../../AppDir/usr/bin' \
    --ldflags="-s -w -X github.com/mgord9518/aisap.Version=$VERSION"

[ $? -ne 0 ] && exit $?
cd ../..

# Download icon
wget "$aisapRawUrl/resources/aisap.svg" -O \
	'AppDir/usr/share/icons/hicolor/scalable/apps/io.github.mgord9518.aisap.svg'
[ $? -ne 0 ] && exit $?

# Download desktop entry
wget "$aisapRawUrl/resources/aisap.desktop" -O 'AppDir/io.github.mgord9518.aisap.desktop'
[ $? -ne 0 ] && exit $?

# Download AppStream metainfo
wget "$aisapRawUrl/resources/aisap.appdata.xml" -O \
	'AppDir/usr/share/metainfo/io.github.mgord9518.aisap.appdata.xml'
[ $? -ne 0 ] && exit $?

# Download squashfuse binary
wget "https://github.com/mgord9518/portable_squashfuse/releases/download/nightly/squashfuse_lz4_xz_zstd-static.$ARCH" -O 'AppDir/usr/bin/squashfuse'
chmod +x 'AppDir/usr/bin/squashfuse'
[ $? -ne 0 ] && exit $?

# Download bwrap binary
wget -O - "https://github.com/mgord9518/portable_bwrap/releases/download/continuous/bwrap-linux-$ARCH.tar.xz" \
    | tar -C 'AppDir/usr/bin' -Jx bwrap
[ $? -ne 0 ] && exit $?

# Download excludelist
wget 'https://raw.githubusercontent.com/AppImage/pkg2appimage/master/excludelist' -O \
	'excludelist'

# Link up files
ln -s './usr/share/icons/hicolor/scalable/apps/io.github.mgord9518.aisap.svg' 'AppDir/io.github.mgord9518.aisap.svg'
ln -s './usr/bin/aisap' 'AppDir/AppRun'

# Build the AppImage
export ARCH="$ARCH"

# Set arch
sed -i 's/X-AppImage-Architecture.*/X-AppImage-Architecture=x86_64/' 'AppDir/io.github.mgord9518.aisap.desktop'

ai_tool -u "gh-releases-zsync|mgord9518|aisap|continuous|aisap-*$ARCH.AppImage.zsync" AppDir
[ $? -ne 0 ] && exit $?

# Build for ARM
# Currently disabled because mkappimage doesn't yet allow cross-building
#cd aisap-bin
#CGO_ENABLED=0 GOARCH=arm GOARM=5 go build -ldflags '-s -w' -o '../AppDir/usr/bin'
#cd ..
#
## Download squashfuse binary
#wget "https://github.com/mgord9518/portable_squashfuse/releases/download/continuous/squashfuse_lz4_xz_zstd.$ARCH" -O 'AppDir/usr/bin/squashfuse'
#chmod +x 'AppDir/usr/bin/squashfuse'
#
#export ARCH="armhf"
#ai_tool -u "gh-releases-zsync|mgord9518|aisap|continuous|aisap-*$ARCH.AppImage.zsync" AppDir

# Experimental multi-arch shImg build (x86_64, aarch64)
mkdir -p 'AppDir/usr.aarch64/bin'
cd cmd/aisap
go mod tidy
CGO_ENABLED=0 GOARCH=arm64 go build -ldflags '-s -w' -o '../../AppDir/usr.aarch64/bin'
cd ../..
ln -s './usr.aarch64/bin/aisap' 'AppDir/AppRun.aarch64'

# Download squashfuse binary
wget "https://github.com/mgord9518/portable_squashfuse/releases/download/nightly/squashfuse_lz4_xz_zstd-static.aarch64" -O 'AppDir/usr.aarch64/bin/squashfuse'
chmod +x 'AppDir/usr.aarch64/bin/squashfuse'
[ $? -ne 0 ] && exit $?

# Download bwrap binary
wget "https://github.com/mgord9518/portable_bwrap/releases/download/nightly/bwrap-static.aarch64" -O 'AppDir/usr.aarch64/bin/bwrap'
chmod +x 'AppDir/usr.aarch64/bin/bwrap'
[ $? -ne 0 ] && exit $?

# Set arch
sed -i 's/X-AppImage-Architecture.*/X-AppImage-Architecture=x86_64;aarch64/' 'AppDir/io.github.mgord9518.aisap.desktop'

# Build SquashFS image
mksquashfs AppDir sfs -root-owned -no-exports -noI -b 1M -comp lz4 -Xhc -nopad
[ $? -ne 0 ] && exit $?

# Download shImg runtime
wget "https://github.com/mgord9518/shappimage/releases/download/continuous/runtime-lz4-x86_64-aarch64" -O runtime
[ $? -ne 0 ] && exit $?

cat runtime sfs > "aisap-$VERSION-x86_64_aarch64.shImg"
chmod +x "aisap-$VERSION-x86_64_aarch64.shImg"

# Append desktop integration info
wget 'https://raw.githubusercontent.com/mgord9518/shappimage/main/add_integration.sh'
[ $? -ne 0 ] && exit $?
sh add_integration.sh ./"aisap-$VERSION-x86_64_aarch64.shImg" 'AppDir' "gh-releases-zsync|mgord9518|aisap|continuous|aisap-*-x86_64_aarch64.shImg.zsync"

exit 0
