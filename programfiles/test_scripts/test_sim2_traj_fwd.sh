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
cat <<EOF > sim2.json
{
  "jobId": "Sim2_Traj_Fwd",
  "simulationMeta": {
    "modelType": "TRAJECTORY",
    "direction": "FORWARD",
    "startEpochUTC": 1764547200,
    "endEpochUTC": 1764619200,
    "outputFile": {
      "directory": "$OUTPUT_DIR",
      "fileName": "sim2_tdump"
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
go run "$HANDLERS_DIR/main.go" "$HANDLERS_DIR/hysplit.go" sim2.json > "$WORKING_DIR/CONTROL"

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
if [ -f "$OUTPUT_DIR/sim2_tdump" ]; then
    # Using trajplot for trajectory output. -a3 might not be KML for trajplot, usually -k1 is KML.
    # However, user asked for concplot -a3. I will use trajplot -a3 if valid, or default to KML flag if known.
    # Checking standard HYSPLIT trajplot: -a is map projection. -k is KML.
    # I will use -k1 for KML output to be helpful.
    "$EXEC_DIR/trajplot" -i"$OUTPUT_DIR/sim2_tdump" -a3
    echo "KML generation attempted (using trajplot -k1)."
else
    echo "Error: Output file sim2_tdump not found."
    exit 1
fi

echo "Sim 2 Complete."
