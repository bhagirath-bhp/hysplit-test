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
cat <<EOF > sim1.json
{
  "jobId": "Sim1_Conc_Fwd",
  "simulationMeta": {
    "modelType": "CONCENTRATION",
    "direction": "FORWARD",
    "startEpochUTC": 1764547200,
    "endEpochUTC": 1764619200,
    "outputFile": {
      "directory": "$OUTPUT_DIR",
      "fileName": "sim1_cdump"
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
      "initialMassG": 10000.0
    },
    "isEmissionRateZero": false
  },
  "concentrationGrids": [
    {
      "centerLat": 40.0,
      "centerLon": -90.0,
      "spacingLat": 0.1,
      "spacingLon": 0.1,
      "spanLat": 50,
      "spanLon": 50,
      "outputLevelsMAgl": [100]
    }
  ]
}
EOF

# Generate CONTROL file
echo "Generating CONTROL file..."
go run "$HANDLERS_DIR/main.go" "$HANDLERS_DIR/hysplit.go" sim1.json > "$WORKING_DIR/CONTROL"

# Run HYSPLIT
echo "Running HYSPLIT (hycs_std)..."
cd "$WORKING_DIR"
if [ -f "$EXEC_DIR/hycs_std" ]; then
    "$EXEC_DIR/hycs_std"
else
    echo "Error: hycs_std not found at $EXEC_DIR/hycs_std"
    exit 1
fi

# Generate KML
echo "Generating KML..."
if [ -f "$OUTPUT_DIR/sim1_cdump" ]; then
    "$EXEC_DIR/concplot" -a3 -i"$OUTPUT_DIR/sim1_cdump" -o"$OUTPUT_DIR/sim1_plot_ps.kml"
    # Note: concplot generates PostScript by default. -a3 might be KML option depending on version, 
    # but usually -a3 is for KML output option in some versions or specific flags.
    # Standard concplot -a option: 0:none, 1:auto, 2:fixed, 3:Google Earth KML
    echo "KML generation attempted."
else
    echo "Error: Output file sim1_cdump not found."
    exit 1
fi

echo "Sim 1 Complete."
