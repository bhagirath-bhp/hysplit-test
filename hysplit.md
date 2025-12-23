```js
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
    "isEmissionRateZero": false // Set to true if the parser should force 0.0 emission rate in CONTROL (e.g., for Sim 3)
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
```

the generic JSON payload structure is **fully capable** of generating the required HYSPLIT `CONTROL` files for all four simulation types (Dispersion FW/BW and Trajectory FW/BW).

The structure leverages flags like `simulationMeta.modelType` and `simulationMeta.direction` to unambiguously define the simulation, allowing the backend logic to conditionally map the data (e.g., using the `points` array as Sources for Forward runs and Receptors for Backward runs).

The critical fields needed for the backward dispersion footprint (`dis.bk`)—like the zero emission rate—are handled by the `pollutantMatrixConfig.isEmissionRateZero` flag, which your backend parser should use to write `0` and `0.0` for the emission rate and duration fields in the `CONTROL` file.

Below are the four separate, minimal JSON payloads for each simulation type, based on your structure and using a consistent 20-hour duration (derived from `startEpochUTC`: 1764547200 and `endEpochUTC`: 1764619200).

---

## 1. Simulation: Dispersion Forward (dis.fw)

**Type:** Concentration-Based Forward Dispersion (Complex Source Term)

| **Change** | **Explanation** |
| --- | --- |
| `direction: "FORWARD"` | The simulation traces pollution *from* the `points` *to* the domain. |
| `emissionScenarios` | **Required** to define the specific time-varying release rates. |
| `isEmissionRateZero: false` | Uses the defined mass/rate in `pollutantMatrixConfig` and `emissionScenarios`. |

```js
{
  "jobId": "Sim1_DisFW",
  "simulationMeta": {
    "modelType": "CONCENTRATION",
    "direction": "FORWARD",
    "startEpochUTC": 1764547200,
    "endEpochUTC": 1764619200
  },
  "metFiles": [
    {"directory": "./", "fileName": "gfs0p25"}
  ],
  "physicsConfig": {
    "configMode": "Particle",
    "maxParticles": 100000,
    "verticalMotionCode": 0,
    "emitimesFilePath": "./EMITIMES", // Used for complex release
    "topOfModelMAgl": 10000.0
  },
  "points": [
    {"pointId": 10, "latitude": 25.1227, "longitude": 55.2385, "heightMAgl": 5.0},
    {"pointId": 14, "latitude": 24.9783, "longitude": 55.202, "heightMAgl": 6.0}
  ],
  "pollutantMatrixConfig": {
    "sox": {"pollutantId": "sox", "initialMassG": 10000.0},
    "nox": {"pollutantId": "nox", "initialMassG": 10000.0},
    "isEmissionRateZero": false
  },
  "concentrationGrids": [
    {"centerLat": 25.0, "centerLon": 55.0, "spacingLat": 0.02, "spacingLon": 0.02, "spanLat": 50, "spanLon": 50, "outputLevelsMAgl": [10.0, 100.0]}
  ],
  "emissionScenarios": [
    // Minimal required emission cycle for the parser to build EMITIMES
    {"pointId": 10, "pollutantId": "sox", "releaseStartEpochUTC": 1764547200, "releaseEndEpochUTC": 1764550800, "rate": {"value": 500.0}},
  ]
}
```

## 2. Simulation: Trajectory Forward (traj.fw)

**Type:** Trajectory-Based Forward Dispersion (Path Tracing)

| **Change** | **Explanation** |
| --- | --- |
| `modelType: "TRAJECTORY"` | Uses the trajectory executable (`hyts_std`). |
| `direction: "FORWARD"` | Traces the path *away* from the starting `points`. |
| Omitted Blocks | `pollutantMatrixConfig`, `concentrationGrids`, and `emissionScenarios` are not used by the trajectory model. |

```js
{
  "jobId": "Sim2_TrajFW",
  "simulationMeta": {
    "modelType": "TRAJECTORY",
    "direction": "FORWARD",
    "startEpochUTC": 1764547200,
    "endEpochUTC": 1764619200
  },
  "metFiles": [
    {"directory": "./", "fileName": "gfs0p25"}
  ],
  "physicsConfig": {
    "verticalMotionCode": 0,
    "topOfModelMAgl": 10000.0
  },
  "points": [
    {"pointId": 10, "latitude": 25.1227, "longitude": 55.2385, "heightMAgl": 5.0},
    {"pointId": 14, "latitude": 24.9783, "longitude": 55.202, "heightMAgl": 6.0}
  ]
}
```

## 3. Simulation: Dispersion Backward (dis.bk)

**Type:** Concentration-Based Backward Dispersion (Footprint / SRS)

| **Change** | **Explanation** |
| --- | --- |
| `direction: "BACKWARD"` | The simulation traces influence *to* the `points` (now Receptors). |
| `pollutantMatrixConfig` | **Required** for the Concentration model. |
| `isEmissionRateZero: true` | **Crucial:** Forces the `CONTROL` file to use $0$ for emission rate/duration, essential for Footprint physics. |
| `startTimeUtcEpoch` | Interpreted as the **Arrival Time** at the receptor. |

```js
{
  "jobId": "Sim3_DisBK",
  "simulationMeta": {
    "modelType": "CONCENTRATION",
    "direction": "BACKWARD",
    "startEpochUTC": 1764547200, // Arrival/Receptor Time
    "endEpochUTC": 1764460800 // 20 hours before arrival (Derived backward end)
  },
  "metFiles": [
    {"directory": "./", "fileName": "gfs0p25"}
  ],
  "physicsConfig": {
    "configMode": "Particle",
    "maxParticles": 100000,
    "verticalMotionCode": 0,
    "topOfModelMAgl": 10000.0
  },
  "points": [
    // These are now the Receptor Locations
    {"pointId": 10, "latitude": 25.1227, "longitude": 55.2385, "heightMAgl": 5.0},
    {"pointId": 14, "latitude": 24.9783, "longitude": 55.202, "heightMAgl": 6.0}
  ],
  "pollutantMatrixConfig": {
    "sox": {"pollutantId": "srs", "initialMassG": 1.0}, // Use a generic ID like 'srs' or 'foot'
    "isEmissionRateZero": true
  },
  "concentrationGrids": [
    // Defines the area over which the footprint/SRS is calculated
    {"centerLat": 25.0, "centerLon": 55.0, "spacingLat": 0.02, "spacingLon": 0.02, "spanLat": 50, "spanLon": 50, "outputLevelsMAgl": [10.0]}
  ],
  "plotConfig": {
    // Requires custom unit for Source Receptor Sensitivity
    "pollutantId": "srs",
    "customUnit": "s/m3", 
    "plotLevelMAgl": 10.0
  }
}
```

## 4. Simulation: Trajectory Backward (traj.bk)

**Type:** Trajectory-Based Backward Dispersion (Back-Trajectory)

| **Change** | **Explanation** |
| --- | --- |
| `modelType: "TRAJECTORY"` | Uses the trajectory executable (`hyts_std`). |
| `direction: "BACKWARD"` | Traces the path *to* the starting time, *away* from the receptor `points`. |
| Omitted Blocks | `pollutantMatrixConfig`, `concentrationGrids`, and `emissionScenarios` are not used. |

```js
{
  "jobId": "Sim4_TrajBK",
  "simulationMeta": {
    "modelType": "TRAJECTORY",
    "direction": "BACKWARD",
    "startEpochUTC": 1764547200, // Arrival/Receptor Time
    "endEpochUTC": 1764460800 // 20 hours before arrival (Derived backward end)
  },
  "metFiles": [
    {"directory": "./", "fileName": "gfs0p25"}
  ],
  "physicsConfig": {
    "verticalMotionCode": 0,
    "topOfModelMAgl": 10000.0
  },
  "points": [
    // These are now the Receptor Locations
    {"pointId": 10, "latitude": 25.1227, "longitude": 55.2385, "heightMAgl": 5.0},
    {"pointId": 14, "latitude": 24.9783, "longitude": 55.202, "heightMAgl": 6.0}
  ]
}
```