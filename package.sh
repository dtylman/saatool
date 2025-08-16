#!/bin/bash

export ANDROID_HOME=$HOME/Android/Sdk
export ANDROID_NDK_HOME=$HOME/Android/Sdk/ndk/29.0.13846066
export PATH=$PATH:$ANDROID_HOME/tools:$ANDROID_HOME/platform-tools

SOURCE_DIR=$(pwd)/cmd/saatool

# ~/go/bin/fyne package --target android/arm64 --app-id org.saatool.app --source-dir $SOURCE_DIR --icon $SOURCE_DIR/icon.png --name "SAATool" 

cd $SOURCE_DIR

go build 

cd ../../

~/go/bin/fyne package --target android/amd64 --app-id org.saatool.app --source-dir $SOURCE_DIR --icon $SOURCE_DIR/icon.png --name "SAAToolEmu" 
