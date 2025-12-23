#!/bin/bash
set -e

# Paths
WORKSPACE_ROOT="/workspaces/hysplit-test"
HANDLERS_DIR="$WORKSPACE_ROOT/handlers"
WORKING_DIR="$WORKSPACE_ROOT/programfiles/hysplit/working"
EXEC_DIR="$WORKSPACE_ROOT/programfiles/hysplit/exec"
OUTPUT_DIR="$WORKSPACE_ROOT/programfiles/output/"
MET_DIR="$WORKSPACE_ROOT/programfiles/metfiles/"

mkdir -p "$OUTPUT_DIR"

# Link ASCDATA.CFG if not present
if [ ! -f "$WORKING_DIR/ASCDATA.CFG" ]; then
    echo "Linking ASCDATA.CFG..."
    ln -sf ../bdyfiles/ASCDATA.CFG "$WORKING_DIR/ASCDATA.CFG"
fi

# JSON Payload
cat <<EOF > sim4.json
{
  "jobId": "Sim4_Traj_Bwd",
  "simulationMeta": {
    "modelType": "TRAJECTORY",
    "direction": "BACKWARD",
    "startEpochUTC": 1764619200,
    "endEpochUTC": 1764547200,
    "outputFile": {
      "directory": "$OUTPUT_DIR",
      "fileName": "sim4_tdump"
    }
  },
  "metFiles": [
    {
      "directory": "$MET_DIR",
      "fileName": "20251201_gfs0p25"
    }
  ],
  "physicsConfig": {
    "verticalMotionCode": 0,
    "topOfModelMAgl": 10000.0
  },
  "points": [
    {
      "pointId": 1,
      "latitude": 40.0,
      "longitude": -90.0,
      "heightMAgl": 100.0
    }
  ],
  "pollutantMatrixConfig": {
    "sox": {
      "pollutantId": "sox",
      "initialMassG": 0.0
    },
    "isEmissionRateZero": false
  },
  "concentrationGrids": []
}
EOF

# Generate CONTROL file
echo "Generating CONTROL file..."
go run "$HANDLERS_DIR/main.go" "$HANDLERS_DIR/hysplit.go" sim4.json > "$WORKING_DIR/CONTROL"

# Run HYSPLIT
echo "Running HYSPLIT (hyts_std)..."
cd "$WORKING_DIR"
if [ -f "$EXEC_DIR/hyts_std" ]; then
    "$EXEC_DIR/hyts_std"
else
    echo "Error: hyts_std not found at $EXEC_DIR/hyts_std"
    exit 1
fi

# Generate KML
echo "Generating KML..."
if [ -f "$OUTPUT_DIR/sim4_tdump" ]; then
    "$EXEC_DIR/trajplot" -i"$OUTPUT_DIR/sim4_tdump"  -a3
    echo "KML generation attempted (using trajplot -a3)."
else
    echo "Error: Output file sim4_tdump not found."
    exit 1
fi

echo "Sim 4 Complete."
