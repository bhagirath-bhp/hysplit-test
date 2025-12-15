#!/bin/bash

# Define variables
HYSPLIT_VERSION="5.4.2"
HYSPLIT_URL="https://www.ready.noaa.gov/data/web/models/hysplit4/linux_trial/hysplit.v${HYSPLIT_VERSION}_x86_64_public.tar.gz"
INSTALL_DIR="/workspaces/hysplit-test/hysplit"

# Create installation directory
mkdir -p "$INSTALL_DIR"

# Download HYSPLIT
echo "Downloading HYSPLIT version $HYSPLIT_VERSION..."
wget -q "$HYSPLIT_URL" -O "$INSTALL_DIR/hysplit.tar.gz"

# Extract the tarball
echo "Extracting HYSPLIT..."
tar -xzf "$INSTALL_DIR/hysplit.tar.gz" -C "$INSTALL_DIR"
# If extraction created a single top-level subdirectory, move its contents up one level
dirs=( "$INSTALL_DIR"/* )
subdirs=()
for p in "${dirs[@]}"; do
    [ -d "$p" ] && subdirs+=( "$p" )
done

if [ "${#subdirs[@]}" -eq 1 ]; then
    src="${subdirs[0]}"
    if [ "$src" != "$INSTALL_DIR" ]; then
        # Move all files (including hidden) from the subdir into INSTALL_DIR
        shopt -s dotglob nullglob
        for item in "$src"/*; do
            mv "$item" "$INSTALL_DIR"/
        done
        shopt -u dotglob nullglob

        # Remove the now-empty subdirectory
        rmdir "$src" 2>/dev/null || rm -rf "$src"
    fi
fi
# Clean up the tarball
rm "$INSTALL_DIR/hysplit.tar.gz"

# Update environment variables
echo "Updating environment variables..."
echo "export HYSPLIT_DIR=$INSTALL_DIR/hysplit" >> "$HOME/.bashrc"
echo "export PATH=\$PATH:\$HYSPLIT_DIR" >> "$HOME/.bashrc"

# Inform the user to source their bashrc
echo "Installation complete. Please run 'source ~/.bashrc' to update your environment."

# End of script
