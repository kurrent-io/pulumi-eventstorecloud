#!/usr/bin/env bash

set -eEo pipefail

# Create a temporary directory to download pulumictl
download_dir=$(mktemp -d)

# Cleanup function to remove the temporary directory
cleanup() {
    local dir="${1}"
    echo "Cleaning up download destination directory ${dir}"
    rm -rf "${dir}"
}

# Set up trap to call cleanup on exit
trap 'cleanup "${download_dir}"' EXIT

# Function to get the latest release tag from pulumi/pulumictl
get_latest_pulumictl_release() {
    local arch="${1:-amd64}"
    local os="${2:-linux}"
    curl -s https://api.github.com/repos/pulumi/pulumictl/releases/latest | \
    jq -r '.assets[] |
         select(.browser_download_url |
         test("pulumictl-.*'"${os}-${arch}"'.tar.gz")) |
         .browser_download_url'
}

# Function to download the latest pulumictl release for a given OS and arch
# Usage: download_pulumictl_release <destination_dir> <os> <arch>
download_pulumictl_release() {
    local dest_dir="$1"
    local arch="${2:-amd64}"
    local os="${3:-linux}"
    local url
    local filename

    filename="pulumictl-${os}-${arch}.tar.gz"

    url=$(get_latest_pulumictl_release "${arch}" "${os}")
    if [[ -z "$url" ]]; then
        echo "No pulumictl release found for ${os} ${arch}"
        return 1
    fi

    mkdir -p "${dest_dir}"
    echo "Downloading pulumictl ${url} for ${os} ${arch}"
    curl -s -L "${url}" -o "${dest_dir}/${filename}"
    echo "Downloaded pulumictl to ${dest_dir}/${filename}"
}

get_os() {
    local os
    os=$(uname -s)
    case $os in
        Linux) echo "linux";;
        Darwin) echo "darwin";;
        *) echo "unknown";;
    esac
}

get_arch() {
    local arch
    arch=$(uname -m)
    case $arch in
        x86_64) echo "amd64";;
        aarch64 | arm64) echo "arm64";;
        *) echo "unknown";;
    esac
}

install_pulumictl() {
    local dest_dir="${1}"
    local os
    local arch
    local filename

    if [[ -z "${dest_dir}" ]]; then
        echo "Destination directory is required"
        return 1
    fi

    os=$(get_os)
    if [[ -z "${os}" ]] || [[ "${os}" == "unknown" ]]; then
        echo "Unknown OS"
        return 1
    fi
    arch=$(get_arch)
    if [[ -z "${arch}" ]] || [[ "${arch}" == "unknown" ]]; then
        echo "Unknown arch"
        return 1
    fi
    download_pulumictl_release "${dest_dir}" "${arch}" "${os}"

    filename="pulumictl-${os}-${arch}.tar.gz"
    echo "Extracting ${filename} to ${dest_dir}"
    tar -xzf "${dest_dir}/${filename}" -C "${dest_dir}"
    sudo mv "${dest_dir}/pulumictl" "/usr/local/bin/pulumictl"
    sudo chmod +x "/usr/local/bin/pulumictl"
    echo "Installed pulumictl to /usr/local/bin/pulumictl"
}

# Install pulumictl
install_pulumictl "${download_dir}"
