const genericPayload = {
  "jobId": "HysplitGenericRun001",
  "simulationMeta": {
    "modelType": "CONCENTRATION",           // **REQUIRED**. Options: "CONCENTRATION" (Sim 1, 3), "TRAJECTORY" (Sim 2, 4)
    "direction": "FORWARD",                  // **REQUIRED**. Options: "FORWARD" (Sim 1, 2), "BACKWARD" (Sim 3, 4)
    "startEpochUTC": 1764547200,      // **REQUIRED**. UNIX epoch time (seconds). // FW: Start of emission. BW: Arrival (Receptor) time.
    "endEpochUTC": 1764619200,           // **REQUIRED**. Future epoch for FORWARD, Past epoch for BACKWARD.
    "outputFile": {
      "directory": "./output/",              // Directory to save output files
      "fileName": "hysplit_output"        // Output file name (e.g., NetCDF format)
    }
  },
  "metFiles": [
    {
      "directory": "./",                        // Path to the meteorological file
      "fileName": "gfs0p25",                  // Your ARL-format GFS file name
    },
  ],
  "physicsConfig": {
    // Maps to SETUP.CFG. Largely constant, but critical for Concentration runs.
    "configMode": "Particle",               // "Particle" or "Puff".
    "maxParticles": 100000,                 // MAXPAR
    "verticalMotionCode": 0,               // Maps to CONTROL line 5 (0: Data, 1: Isobaric, etc.)
    "emitimesFilePath": "./EMITIMES",       // REQUIRED for Sim 1 (Concentration Forward). Omit for all others.
    "topOfModelMAgl": 10000.0,             // Top of model domain (m AGL)
  },
  "points": [
    // **REQUIRED for all simulations**. Interpreted based on 'direction':
    // FORWARD: Source Location(s)
    // BACKWARD: Receptor Location(s) (Arrival/Sampling Points)
    {
      "pointId": 10,
      "latitude": 25.1227,
      "longitude": 55.2385,
      "heightMAgl": 5.0
    },
    {
      "pointId": 14,
      "latitude": 24.9783,
      "longitude": 55.202,
      "heightMAgl": 6.0
    }
  ],
  "units": {
    "u1": {
      "unitId": "u1",
      "label": "g/m3",
      "description": "grams per cubic meter",
      "conversion_strategy": {
        "m": 0.0019631,
        "type": "mx"
      },
      "custom_zones": {
        "next": [
          {
            "color": "#6ecc58",
            "upper": 1,
            "inverted_color": "#000000"
          },
          {
            "color": "#bbcf4c",
            "upper": 2,
            "inverted_color": "#000000"
          }
        ],
        "lower": 0
      },
    }
  },
  "pollutantMatrixConfig": {
    // **REQUIRED for Concentration runs (Sim 1, 3)**. Used to define the pollutant ID matrix in the CONTROL file.
    // Omit or leave empty for Trajectory runs (Sim 2, 4).
    // "pollutants": [
    //     {"id": "SOX", "initialMassG": 10000.0},
    //     {"id": "NOX", "initialMassG": 10000.0}
    // ],
    "sox": {
      "pollutantId": "sox",
      "initialMassG": 10000.0,
      "unitId": "u1"
    },
    "nox": {
      "pollutantId": "nox",
      "initialMassG": 10000.0,
      "unitId": "u1"
    },
    // "isEmissionRateZero": false // Set to true if the parser should force 0.0 emission rate in CONTROL (e.g., for Sim 3)
    // **For Concentration Backward (Sim 3)**, the parser must use 0.0 for CONTROL file Emission Rate/Duration.
  },

  "concentrationGrids": [
      {
      // **REQUIRED for Concentration runs (Sim 1, 3)**. Omit for Trajectory runs.
      "centerLat": 25.0,
      "centerLon": 55.0,
      "spacingLat": 0.02,
      "spacingLon": 0.02,
      "spanLat": 50,
      "spanLon": 50,
      "outputLevelsMAgl": [10.0, 100.0] // Array of height levels for output
    },
  ],
  "emissionScenarios": [
    // **REQUIRED ONLY for Complex Concentration Forward (Sim 1)**. Omit for all others.
    // The parser must ensure a record for every source/pollutant combination exists in each cycle.
    {"pointId": 10, "pollutantId": "sox", "releaseStartEpochUTC": 1743993000, "releaseEndEpochUTC": 1743994200, "rate": {"value": 500.0, "unitId": "u1"}, "area": {"value": 100.0, "unitId": "m2"}},
    {"pointId": 10, "pollutantId": "nox", "releaseStartEpochUTC": 1743993000, "releaseEndEpochUTC": 1743994200, "rate": {"value": 300.0, "unitId": "u1"}, "area": {"value": 100.0, "unitId": "m2"}},
  ],

  "plotConfig": {
    "pollutantId": "sox",
    "plotLevelMAgl": 10.0,
    "outputPlotFile": {
      "directory": "./plots/",
      "fileName": "concentration_plot.png"
    },
    "contourLevels": [
      {"value": 1.0, "colorIndex": 5}
    ]
  }
}