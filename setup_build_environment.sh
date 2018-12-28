#!/bin/bash

# Source this file by running 'source setup_build_environment.sh' (from the directory
# this file is checked in). This will set your config for building and running
# OurJamz client/server stack.

usage()
{
	echo Usage: source ./setup_environment.sh [--nopkg] [--novs] [--env]
	echo '       --env:   Only set environment, skip everything else'
	echo '       --nopkg: Do not retrieve packages'
	echo '       --novs:  Do not launch Visual Studio code'
	echo Note: Must be run in the $EXPECTED_DIRECTORY directory
}

parse_args()
{
	while [ $# -ne 0 ]
	do
		opt="$1"
		case "$opt" in
			h|-h|-help|--help|/?)
				usage
				EXIT_NOW=1
       			;;
			-nopkg|--nopkg)
				NO_PACKAGE=1
			;;
			-novs|--novs)
				NO_VS=1
       			;;
			-env|--env)
				NO_VS=1
				NO_PACKAGE=1
       			;;
		esac
		shift
	done
}

launch_vs_code()
{
	# Launch Visual Studio Code.
	uname=`uname`
	if [[ ${uname} == 'Darwin' ]]; then
    		# Launch 'Visual Studio Code' on Mac.
    		cmd=/Applications/Visual\ Studio\ Code.app/Contents/MacOS/Electron
	elif [[ ${uname} == 'Linux' ]]; then
    		# Launch Visual Studio code on Linux.
    		cmd=/usr/share/code/code
	elif [[ ${uname} == MING* ]]; then
    		# Launch Visual Studio code on Linux.
		cmd=code
	elif [[ ${uname} == CYGWIN_NT* ]]; then
		cmd=code
	fi
	echo Launching Visual Studio Code
	"${cmd}" . &
}

clear_env()
{
	unset EXIT_NOW
	unset NO_PACKAGE
	unset NO_VS
	unset NO_PACKAGE
	unset EXPECTED_DIRECTORY
}

show_env()
{
	echo EXPECTED_DIRECTORY=$EXPECTED_DIRECTORY
	echo EXIT_NOW=$EXIT_NOW
	echo NO_PACKAGE=$NO_PACKAGE
	echo NO_VS=$NO_VS
	echo NO_PACKAGE=$NO_PACKAGE
}

clear_env
EXPECTED_DIRECTORY="go4004"
parse_args "$@"
if [[ $EXIT_NOW -eq 1 ]]; then
	clear_env
	return 0
fi

# Make sure we are in the go4004 directory
uname=`uname`
if [[ ${uname} == CYGWIN_NT* || ${uname} == MSYS_NT* ]]; then
	CURRDIR=`pwd`
	CURRDIR=`cygpath.exe -w $CURRDIR`
else
	CURRDIR=`pwd`
fi


echo Running from $CURRDIR
if ! [[ $CURRDIR =~ ^.*$EXPECTED_DIRECTORY$ ]]; then
	echo ERROR: Unexpected directory
	usage
	clear_env
	return 1
fi

export GOPATH=$CURRDIR
echo "GOPATH is set to $GOPATH"

if [[ $NO_PACKAGE -ne 1 ]]; then
	# Get all 'Go' packages the server code depends on.
	source ./packages.sh
fi

# Set path to MingW gcc and temporary directory.
if [[ ${uname} == CYGWIN_NT* || ${uname} == MING*  || ${uname} == MSYS_NT* ]]; then
	# export PATH=$PATH:/c/cygwin64/bin
	echo Setting up FFMPEG for Windows
	# No don't do this, it leads to a path that grows indefinitely.
	# User should set this up ahead of time based on docuemtnation in Wiki
	# export PATH=$PATH:/c/Mingw/mingw64/bin:/c/msys64/usr/bin
	export TEMP="C:\\TEMP"
	export PKG_CONFIG_PATH=$CURRDIR/src/ffmpeglib/windows:$CURRDIR/src/ffmpeglib/
	./ffmpeg_setup_lib.sh
	mkdir -p $TEMP
elif [[ ${uname} == 'Linux' ]]; then
	echo Setting up FFMPEG for Linux
	# Anything to do here?
else
# for Mac's that have ffmpeg installed via Brew. YMMV!
	echo Setting up FFMPEG for Mac
	export PKG_CONFIG_PATH=/usr/local/Cellar/ffmpeg/3.4/lib/pkgconfig
fi

if [[ ${uname} == CYGWIN_NT* || ${uname} == MING*  || ${uname} == MSYS_NT* || ${uname} == Darwin ]]; then
	./copy_dlls.sh
fi

if [[ $NO_VS -ne 1 ]]; then
	launch_vs_code
fi
clear_env

return 0
