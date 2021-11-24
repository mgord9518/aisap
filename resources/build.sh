#!/bin/sh
# Quick and dirty shell script to build aisap AppDir from source

[ -z "$ARCH" ] && ARCH=$(uname -m)

if command -v 'appimagetool.AppImage'; then
	aitool() {
		appimagetool.AppImage "$@"
	}
elif command -v "appimagetool-$ARCH.AppImage"; then
	aitool() {
		"appimagetool-$ARCH.AppImage" "$@"
	}
elif command -v "appimagetool"; then
	aitool() {
		"appimagetool" "$@"
	}
else
	echo 'Failed to locate appimagetool in $PATH! Unable to build'
	exit 1
fi

if command -v 'go'; then
	gocc() {
		go "$@"
	}
elif command -v 'gccgo'; then
	gocc() {
		go "$@"
	}
else
	echo 'Failed to locate GoLang compiler! Unable to build'
	exit 1
fi

aisapUrl='github.com/mgord9518/aisap'
aisapRawUrl='raw.githubusercontent.com/mgord9518/aisap/main'

mkdir -p 'AppDir/usr/bin' \
         'AppDir/usr/share/icons/hicolor/scalable/apps'

# Download and compile the binary in the current directory
GOBIN="$PWD/AppDir/usr/bin" gocc install -ldflags '-s -w' \
	"$aisapUrl/aisap-bin@latest"

# Download icon
wget "$aisapRawUrl/resources/aisap.svg" -O \
	'AppDir/usr/share/icons/hicolor/scalable/apps/aisap.svg'

# Download desktop entry
wget "$aisapRawUrl/resources/aisap.desktop" -O 'AppDir/aisap.desktop'

# Download squashfuse
wget "$aisapRawUrl/resources/squashfuse" -O 'AppDir/usr/bin/squashfuse'
chmod +x 'AppDir/usr/bin/squashfuse'

# Link up files
ln -s './usr/share/icons/hicolor/scalable/apps/aisap.svg' 'AppDir/aisap.svg'
ln -s './usr/bin/aisap-bin' 'AppDir/AppRun'

# Build the AppImage
ARCH="$ARCH" aitool AppDir
