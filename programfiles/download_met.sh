#!/bin/bash
set -euo pipefail

# Enable showing executed commands (optional, comment out if noisy)
# set -x

# Check if a date string argument is provided
if [[ -z "${1:-}" ]]; then
    echo "Usage: $0 <date_str>"
    exit 1
fi

DATE_STR="$1"
MET_DIR="/workspaces/hysplit-test/metfiles"  # Change this to the desired MET_DIR
FTP_HOST="ftp.arl.noaa.gov"    # Change this to the actual FTP host
FTP_DIR="/archives/gfs0p25"     # Change this to the actual FTP directory
MAX_MET_FILES_SIZE_BYTES=26843545600  # Set the max size in bytes (e.g., 25 GB)

PYTHON_SCRIPT_PATH="/tmp/download_met_file.py"

# Write the Python script (variables expanded so no sed replacement required)
cat > "$PYTHON_SCRIPT_PATH" <<'PY'
#!/usr/bin/env python3
import os
import ftplib
import logging
import sys
from typing import Dict

# Configure logging to stdout and enable DEBUG for verbose subprocess/ftplib info
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s: %(message)s",
    stream=sys.stdout,
)
logger = logging.getLogger(__name__)

def download_met_file(date_str: str, config: Dict[str, object]) -> bool:
    from contextlib import contextmanager
    import time

    class Timer:
        def __init__(self, message):
            self.message = message

        def __enter__(self):
            self.start = time.time()
            return self

        def __exit__(self, *args):
            end = time.time()
            logger.info(f"{self.message} took {end - self.start:.2f} seconds")

    with Timer(f"Download met file {date_str}"):
        filename = f"{date_str}_gfs0p25"
        local_path = os.path.join(config['MET_DIR'], filename)

        # Ensure MET_DIR exists
        os.makedirs(config['MET_DIR'], exist_ok=True)

        # Check size of MET_DIR
        total_size = 0
        oldest_file = None
        for dirpath, dirnames, filenames in os.walk(config['MET_DIR']):
            for f in filenames:
                fp = os.path.join(dirpath, f)
                try:
                    if oldest_file is None or os.path.getctime(fp) < os.path.getctime(oldest_file):
                        oldest_file = fp
                    total_size += os.path.getsize(fp)
                except OSError:
                    logger.exception("Error accessing file stats for %s", fp)

        logger.info("Current MET_DIR size: %d bytes (limit %d)", total_size, config['MAX_MET_FILES_SIZE_BYTES'])
        if total_size >= config['MAX_MET_FILES_SIZE_BYTES']:
            if oldest_file:
                try:
                    logger.info("Removing oldest file to free space: %s", oldest_file)
                    os.remove(oldest_file)
                except OSError:
                    logger.exception("Failed to remove oldest file %s", oldest_file)

        # Check if already exists
        if os.path.exists(local_path) and os.path.getsize(local_path) > 0:
            logger.info("Met file %s already exists", filename)
            return True

        try:
            # Connect and enable ftplib debug output
            with ftplib.FTP(config['FTP_HOST']) as ftp:
                ftp.set_debuglevel(2)  # <-- enables verbose FTP protocol exchange on stdout
                logger.info("Connecting to FTP: %s", config['FTP_HOST'])
                ftp.login()
                ftp.cwd(config['FTP_DIR'])

                # Download file
                logger.info("📥 Downloading %s ...", filename)
                with open(local_path, 'wb') as f:
                    ftp.retrbinary(f'RETR {filename}', f.write)

                if os.path.exists(local_path) and os.path.getsize(local_path) > 0:
                    logger.info("✅ Downloaded %s", filename)
                    return True
                else:
                    logger.error("❌ Download failed - file is empty")
                    return False

        except Exception:
            logger.exception("❌ Error downloading %s", filename)
            return False

if __name__ == "__main__":
    # expanded by the calling shell; placeholder values are replaced below in the bash script
    config = {
        'MET_DIR': os.environ.get('MET_DIR', '/path/to/metfiles'),
        'FTP_HOST': os.environ.get('FTP_HOST', 'ftp.example.com'),
        'FTP_DIR': os.environ.get('FTP_DIR', '/path/to/ftp/dir'),
        'MAX_MET_FILES_SIZE_BYTES': int(os.environ.get('MAX_MET_FILES_SIZE_BYTES', '26843545600')),
    }

    # date string passed as an env var by the caller
    date_str = os.environ.get('DATE_STR', '')
    if not date_str:
        logger.error("DATE_STR environment variable is required")
        sys.exit(2)

    success = download_met_file(date_str, config)
    sys.exit(0 if success else 1)
PY

chmod +x "$PYTHON_SCRIPT_PATH"

# Export env so Python script picks up config and runs unbuffered for immediate logs
export MET_DIR FTP_HOST FTP_DIR MAX_MET_FILES_SIZE_BYTES DATE_STR="$DATE_STR"
export PYTHONUNBUFFERED=1

# Use a virtualenv optionally; not required for ftplib (it's stdlib). If you do want venv, uncomment below.
# python3 -m venv /tmp/met_env
# source /tmp/met_env/bin/activate

# Run the script with unbuffered output so logs appear in terminal in real time
# Use python3 -u explicitly to force unbuffered binary stdout/stderr
python3 -u "$PYTHON_SCRIPT_PATH" 2>&1 | tee /dev/stderr

# Cleanup
# deactivate 2>/dev/null || true
rm -f "$PYTHON_SCRIPT_PATH"

echo "Process completed."