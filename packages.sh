#!/bin/sh

# Run this script to download all packages needed by Go server.

echo Getting required packages for `pwd`
#packages to download and build with go
gopackages='
    github.com/derekparker/delve/cmd/dlv
    github.com/fogleman/gg
	github.com/tfriedel6/canvas/glfwcanvas
	github.com/romana/rlog
'

linux_gopackages=''

linux_ud_gopackages=''

windows_gopackages=''

# List of Go thirdparty packages which will always be pulled with the '-u' (update flag)
gopackages_always_update=''

# Determine the go executable name based on platform
uname=`uname`
if [[ (${uname} = CYGWIN_NT*) || (${uname} == MING*) ]]; then
	GOCMD="go.exe"
else
    GOCMD="go"
fi

# Download packages but no updates are applied
for package in ${gopackages}
do
    echo '  Getting' $package
    $GOCMD get $package
done

# Download packages also apply updates
for package in ${gopackages_always_update}
do
    echo '  Getting' $package
    $GOCMD get -u $package
done

if [[ (${uname} = CYGWIN_NT*) || (${uname} == MING*) ]]; then
	for package in ${windows_gopackages}
	do
		echo '  Getting' $package
		$GOCMD get $package
	done
fi

if [[ ${uname} = Linux ]]; then
	for package in ${linux_gopackages}
	do
		echo '  Getting' $package
		$GOCMD get $package
	done
	for package in ${linux_ud_gopackages}
	do
		echo '  Getting' $package
		$GOCMD get -u -d $package
		echo NOTE: You may need to run the following commands
		echo pushd src/$package
		echo make install
		echo popd
	done
fi

echo Done
