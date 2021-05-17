#!/bin/bash
# build jire-release-maker executables for different platforms

build_version="0.0.1"
platforms=("linux/amd64" "linux/arm64" "linux/arm" "darwin/amd64" "darwin/arm64" "windows/amd64")

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name='builds/jira-release-maker-'$build_version'-'$GOOS'-'$GOARCH
    if [ $GOOS = "windows" ]; then
        output_name+=".exe"
    fi

    env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name cmd/cli/*.go
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done
