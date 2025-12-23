package main

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// Payload is the root structure for the generic JSON input.
type Payload struct {
	JobId             string            `json:"jobId"`
	SimulationMeta    SimulationMeta    `json:"simulationMeta"`
	MetFiles          []MetFile         `json:"metFiles"`
	PhysicsConfig     PhysicsConfig     `json:"physicsConfig"`
	Points            []Point           `json:"points"`
	PollutantMatrixConfig PollutantMatrixConfig `json:"pollutantMatrixConfig"`
	ConcentrationGrids []ConcentrationGrid `json:"concentrationGrids"`
}

type SimulationMeta struct {
	ModelType string `json:"modelType"` // "CONCENTRATION" or "TRAJECTORY"
	Direction string `json:"direction"` // "FORWARD" or "BACKWARD"
	StartEpochUTC int64 `json:"startEpochUTC"`
	EndEpochUTC   int64 `json:"endEpochUTC"`
	OutputFile    OutputMeta `json:"outputFile"`
}

type OutputMeta struct {
	Directory string `json:"directory"`
	FileName  string `json:"fileName"`
}

type MetFile struct {
	Directory string `json:"directory"`
	FileName  string `json:"fileName"`
}

type PhysicsConfig struct {
	VerticalMotionCode int `json:"verticalMotionCode"`
	TopOfModelMAgl     float64 `json:"topOfModelMAgl"`
}

type Point struct {
	PointId  int     `json:"pointId"`
	Latitude float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	HeightMAgl float64 `json:"heightMAgl"`
}

type PollutantMatrixConfig struct {
	SOX struct {
		PollutantId  string `json:"pollutantId"`
		InitialMassG float64 `json:"initialMassG"`
	} `json:"sox"`
	IsEmissionRateZero bool `json:"isEmissionRateZero"`
}

type ConcentrationGrid struct {
	CenterLat        float64   `json:"centerLat"`
	CenterLon        float64   `json:"centerLon"`
	SpacingLat       float64   `json:"spacingLat"`
	SpacingLon       float64   `json:"spacingLon"`
	SpanLat          int       `json:"spanLat"` // Number of grid points
	SpanLon          int       `json:"spanLon"`
	OutputLevelsMAgl []float64 `json:"outputLevelsMAgl"`
}

// ------------------- Helper Functions -------------------

// epochToHysplitTime converts a Unix epoch to the HYSPLIT format "YY MM DD HH MM".
func epochToHysplitTime(epoch int64) string {
	t := time.Unix(epoch, 0).UTC()
	return fmt.Sprintf("%02d %02d %02d %02d %01d",
		t.Year()%100, // YY
		t.Month(),    // MM
		t.Day(),      // DD
		t.Hour(),     // HH
		t.Minute(),   // M (single digit for minutes, usually 0)
	)
}

// calculateDuration determines the signed duration in hours.
func calculateDuration(meta SimulationMeta) float64 {
	durationSeconds := meta.EndEpochUTC - meta.StartEpochUTC
	durationHours := float64(durationSeconds) / 3600.0

	if meta.Direction == "BACKWARD" {
		return math.Abs(durationHours) * -1.0
	}
	return math.Abs(durationHours)
}

// ------------------- Main Generator Function -------------------

// func GenerateHysplitControlFile(payload Payload) (string, error) {
// 	meta := payload.SimulationMeta
// 	met := payload.MetFiles[0] // Assuming one met file
// 	phys := payload.PhysicsConfig
// 	points := payload.Points

// 	var sb strings.Builder

// 	// 1. STARTING YEAR MONTH DAY HOUR MINUTES
// 	// This is the release time for FW, or the arrival time for BW.
// 	sb.WriteString(fmt.Sprintf("%s\t\t\t#STARTING YEAR MONTH DAY HOUR MINUTES\n", epochToHysplitTime(meta.StartEpochUTC)))

// 	// 2. NUMBER OF SOURCE LOCATIONS
// 	sb.WriteString(fmt.Sprintf("%d\t\t\t#NUMBER OF SOURCE LOCATIONS\n", len(points)))

// 	// 3. SOURCE 1 LATITUDE LONGITUDE HEIGHT(m-agl) (Repeated for all points)
// 	for i, p := range points {
// 		// Use tab for consistent spacing observed in HYSPLIT output, though space is usually sufficient.
// 		sb.WriteString(fmt.Sprintf("%f %f %0.2f\t\t#SOURCE %d LATITUDE LONGITUDE HEIGHT(m-agl)\n",
// 			p.Latitude, p.Longitude, p.HeightMAgl, i+1))
// 	}

// 	// 4. TOTAL RUN TIME (backwards/forwards)
// 	duration := calculateDuration(meta)
// 	sb.WriteString(fmt.Sprintf("%0.f\t\t\t#TOTAL RUN TIME (hours)\n", duration))

// 	// 5. USE MODEL VERTICAL VELOCITY
// 	sb.WriteString(fmt.Sprintf("%d\t\t\t#USE MODEL VERTICAL VELOCITY\n", phys.VerticalMotionCode))

// 	// 6. TOP OF MODEL DOMAIN (m-AGL)
// 	sb.WriteString(fmt.Sprintf("%0.f\t\t\t#TOP OF MODEL DOMAIN (m-AGL)\n", phys.TopOfModelMAgl))

// 	// 7. NUMBER OF INPUT DATA GRIDS
// 	sb.WriteString(fmt.Sprintf("1\t\t\t\t#NUMBER nextfile mfile OF INPUT DATA GRIDS\n"))

// 	// 8. MET FILE DIRECTORY
// 	sb.WriteString(fmt.Sprintf("%s\t\t\t#MET FILE DIRECTORY\n", met.Directory))

// 	// 9. MET FILE NAME
// 	sb.WriteString(fmt.Sprintf("%s\t\t\t#MET FILE NAME\n", met.FileName))

// 	// --- Conditional Blocks based on Model Type ---

// 	if meta.ModelType == "CONCENTRATION" {
// 		// --- Concentration Specific Block (Sim 1, Sim 3) ---

// 		grid := payload.ConcentrationGrids[0] // Assuming one grid

// 		// 10. NUMBER OF DIFFERENT POLLUTANTS
// 		// We only define one pollutant here based on your minimal matrix config.
// 		sb.WriteString(fmt.Sprintf("1\t\t\t\t#NUMBER OF DIFFERENT POLLUTANTS\n"))

// 		// 11. POLLUTANT IDENTIFICATION and MASS (For Concentration only)
// 		pollutantID := strings.ToUpper(payload.PollutantMatrixConfig.SOX.PollutantId)
// 		massG := payload.PollutantMatrixConfig.SOX.InitialMassG
// 		if meta.Direction == "BACKWARD" {
// 			// For backward footprint (Sim 3), HYSPLIT often expects a dummy mass (e.g., 1.0 or 0.0)
// 			// and 'Unit' or 'SRS' as the ID. We use the configured ID but force mass to 1.0 or 0.0.
// 			massG = 0.0
// 			sb.WriteString(fmt.Sprintf("%0.1f %s\t\t\t#POLLUTANT IDENTIFICATION AND INITIAL MASS\n", massG, pollutantID))
// 		} else {
// 			// For forward concentration (Sim 1), use the defined mass.
// 			sb.WriteString(fmt.Sprintf("%0.1f %s\t\t\t#POLLUTANT IDENTIFICATION AND INITIAL MASS\n", massG, pollutantID))
// 		}


// 		// 12. EMISSION RATE (per hour)
// 		// 13. HOURS OF EMISSION
// 		emissionRate := 0.0
// 		emissionHours := 0.0
// 		releaseTime := "00 00 00 00 0"

// 		if meta.Direction == "FORWARD" {
// 			// For Complex Forward (Sim 1), the rate/duration in CONTROL must be non-zero (or 0 if EMITIMES is used).
// 			// We set a simple default or derive from the first scenario for CONTROL, and rely on EMITIMES.
// 			emissionRate = 1.0 // Placeholder
// 			emissionHours = 1.0 // Placeholder
// 			releaseTime = epochToHysplitTime(meta.StartEpochUTC)
// 		} else if meta.Direction == "BACKWARD" {
// 			// CRUCIAL for Backward Footprint (Sim 3): must be 0/0.0
// 			emissionRate = 0.0
// 			emissionHours = 0.0
// 			releaseTime = "00 00 00 00 0"
// 		}

// 		sb.WriteString(fmt.Sprintf("%0.f\t\t\t\t#EMISSION RATE (per hour)\n", emissionRate))
// 		sb.WriteString(fmt.Sprintf("%0.1f \t\t\t#HOURS OF EMISSION\n", emissionHours))

// 		// 14. RELEASE START TIME:YEAR MONTH DAY HOUR MINUTE (Must be 00... for Footprint)
// 		sb.WriteString(fmt.Sprintf("%s \t\t#RELEASE START TIME:YEAR MONTH DAY HOUR MINUTE\n", releaseTime))

// 		// 15. NUMBER OF CONCENTRATION GRIDS
// 		sb.WriteString(fmt.Sprintf("1\t\t\t\t#NUMBER OF CONCENTRATION GRIDS\n"))

// 		// 16. CONC GRID CENTER LATITUDE LONGITUDE
// 		sb.WriteString(fmt.Sprintf("%0.1f %0.1f \t\t\t#CONC GRID CENTER LATITUDE LONGITUDE\n", grid.CenterLat, grid.CenterLon))

// 		// 17. CONC GRID SPACING (degrees) LATITUDE LONGITUDE
// 		sb.WriteString(fmt.Sprintf("%0.3f %0.3f\t\t#CONC GRID SPACING (degrees) LATITUDE LONGITUDE\n", grid.SpacingLat, grid.SpacingLon))

// 		// 18. CONC GRID SPAN (NUMBER OF POINTS)
// 		// Note: The example file used 0.8 0.8 which is incorrect (should be integers). We use the SpanLat/Lon integers.
// 		sb.WriteString(fmt.Sprintf("%d %d\t\t\t#CONC GRID SPAN (NUMBER OF POINTS)\n", grid.SpanLat, grid.SpanLon))

// 		// 19. OUTPUT DIRECTORY
// 		sb.WriteString(fmt.Sprintf("%s\t\t\t#OUTPUT DIRECTORY\n", meta.OutputFile.Directory))

// 		// 20. OUTPUT FILENAME
// 		sb.WriteString(fmt.Sprintf("%s\t\t\t#OUTPUT FILENAME\n", meta.OutputFile.FileName))

// 		// 21. NUMBER OF VERTICAL CONCENTRATION LEVELS
// 		sb.WriteString(fmt.Sprintf("%d\t\t\t\t#NUMBER OF VERTICAL CONCENTRATION LEVELS\n", len(grid.OutputLevelsMAgl)))

// 		// 22. HEIGHT OF EACH CONCENTRATION LEVEL (m-agl)
// 		levelStrings := make([]string, len(grid.OutputLevelsMAgl))
// 		for i, level := range grid.OutputLevelsMAgl {
// 			levelStrings[i] = fmt.Sprintf("%0.f", level)
// 		}
// 		sb.WriteString(fmt.Sprintf("%s\t\t\t#HEIGHT OF EACH CONCENTRATION LEVEL (m-agl)\n", strings.Join(levelStrings, " ")))

// 		// 23. SAMPLING START TIME: YEAR MONTH DAY HOUR MINUTE
// 		// For backward, this is crucial and matches line 1 (arrival time).
// 		sb.WriteString(fmt.Sprintf("%s \t\t#SAMPLING START TIME:YEAR MONTH DAY HOUR MINUTE\n", epochToHysplitTime(meta.StartEpochUTC)))

// 		// 24. SAMPLING STOP TIME: YEAR MONTH DAY HOUR MINUTE (Relative time)
// 		// For a run that calculates the total footprint/plume over the duration, this is often set to the total duration.
// 		// Using the relative time format from the example: 00 00 00 12 00 (relative 12 hours)
// 		sb.WriteString(fmt.Sprintf("00 00 00 %02d 00\t\t#SAMPLING STOP TIME: YEAR MONTH DAY HOUR MINUTE (Relative duration)\n", int(math.Abs(duration))))

// 		// 25. SAMPLING INTERVAL: TYPE HOUR MINUTE (0 1 0 = every 1 hour)
// 		sb.WriteString(fmt.Sprintf("0 1 0\t\t\t\t#SAMPLING INTERVAL: TYPE HOUR MINUTE\n"))

// 		// 26. DEPOSITION/CHEMISTRY FIELDS (Standard, non-depositing values for Footprint/Simple Conc)
// 		sb.WriteString("1\t\t\t\t#NUMBER OF DEPOSITING POLLUTANTS\n")
// 		sb.WriteString("0.0 0.0 0.0\t\t\t\t#PARTICLE:DIAMETER (um), DENSITY (g/cc), SHAPE\n")
// 		sb.WriteString("0.0 0.0 0.0 0.0 0.0\t\t\t#DEP VEL (m/s), MW (g/Mole), SFC REACT. RATIO, DIFFUSIVITY RATIO, HENRY'S CONSTANT\n")
// 		sb.WriteString("0.0 0.0 0.0\t\t\t#WET REMOVAL: HENRY'S (Molar/atm), IN-CLOUD (1/s), BELOW-CLOUD (1/s)\n")
// 		sb.WriteString("0\t\t\t\t#RADIOACTIVE DECAY HALF-LIFE (days)\n")
// 		sb.WriteString("0.0\t\t\t\t#POLLUTANT RESUSPENSION (1/m)\n")


// 	} else if meta.ModelType == "TRAJECTORY" {
// 		// --- Trajectory Specific Block (Sim 2, Sim 4) ---

// 		// These are the two additional fields that must be present in the Trajectory CONTROL file.
// 		// They control the frequency of output for the trajectory points.

// 		// 10. Trajectory Plot Interval (Number of time steps/minutes)
// 		sb.WriteString("0\t\t\t\t#TRAJECTORY PLOT INTERVAL (minutes)\n")

// 		// 11. Vertical Levels (Number of levels to plot)
// 		sb.WriteString("100\t\t\t\t#VERTICAL LEVELS FOR PLOTTING\n")
// 	}

// 	return sb.String(), nil
// }

func GenerateHysplitControlFile(payload Payload) (string, error) {
	meta := payload.SimulationMeta
	phys := payload.PhysicsConfig
	points := payload.Points
	metfiles := payload.MetFiles

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s\n", epochToHysplitTime(meta.StartEpochUTC)))

	sb.WriteString(fmt.Sprintf("%d\n", len(points)))

	for _, p := range points {
		sb.WriteString(fmt.Sprintf("%f %f %0.2f\n",
			p.Latitude, p.Longitude, p.HeightMAgl))
	}

	duration := calculateDuration(meta)
	sb.WriteString(fmt.Sprintf("%0.f\n", duration))

	sb.WriteString(fmt.Sprintf("%d\n", phys.VerticalMotionCode))

	sb.WriteString(fmt.Sprintf("%0.f\n", phys.TopOfModelMAgl))

	sb.WriteString(fmt.Sprintf("%d\n", len(metfiles)))

	for _, mf := range metfiles {
		sb.WriteString(fmt.Sprintf("%s\n", mf.Directory))
		sb.WriteString(fmt.Sprintf("%s\n", mf.FileName))
	}
	if meta.ModelType == "CONCENTRATION" {
		grid := payload.ConcentrationGrids[0]

		sb.WriteString(fmt.Sprintf("1\n"))

		pollutantID := strings.ToUpper(payload.PollutantMatrixConfig.SOX.PollutantId)
		if meta.Direction == "BACKWARD" {
			sb.WriteString(fmt.Sprintf("%s\n", pollutantID))
		} else {
			sb.WriteString(fmt.Sprintf("%s\n", pollutantID))
		}

		emissionRate := 1.0
		emissionHours := 1.0
		releaseTime := "00 00 00 00 0"

		if meta.Direction == "FORWARD" {
			emissionRate = 1.0
			emissionHours = 1.0
			releaseTime = epochToHysplitTime(meta.StartEpochUTC)
		} 
		// else if meta.Direction == "BACKWARD" {
		// 	emissionRate = 0.0
		// 	emissionHours = 0.0
		// 	releaseTime = "00 00 00 00 0"
		// }

		sb.WriteString(fmt.Sprintf("%0.f\n", emissionRate))
		sb.WriteString(fmt.Sprintf("%0.1f\n", emissionHours))

		sb.WriteString(fmt.Sprintf("%s\n", releaseTime))

		sb.WriteString(fmt.Sprintf("1\n"))

		sb.WriteString(fmt.Sprintf("%0.1f %0.1f\n", grid.CenterLat, grid.CenterLon))

		sb.WriteString(fmt.Sprintf("%0.3f %0.3f\n", grid.SpacingLat, grid.SpacingLon))

		sb.WriteString(fmt.Sprintf("%d %d\n", grid.SpanLat, grid.SpanLon))

		sb.WriteString(fmt.Sprintf("%s\n", meta.OutputFile.Directory))

		sb.WriteString(fmt.Sprintf("%s\n", meta.OutputFile.FileName))

		sb.WriteString(fmt.Sprintf("%d\n", len(grid.OutputLevelsMAgl)))

		levelStrings := make([]string, len(grid.OutputLevelsMAgl))
		for i, level := range grid.OutputLevelsMAgl {
			levelStrings[i] = fmt.Sprintf("%0.f", level)
		}
		sb.WriteString(fmt.Sprintf("%s\n", strings.Join(levelStrings, " ")))

		sb.WriteString(fmt.Sprintf("%s\n", epochToHysplitTime(meta.StartEpochUTC)))

		sb.WriteString(fmt.Sprintf("00 00 00 %02d 00\n", int(math.Abs(duration))))

		sb.WriteString(fmt.Sprintf("0 1 0\n"))

		sb.WriteString("1\n")
		sb.WriteString("0.0 0.0 0.0\n")
		sb.WriteString("0.0 0.0 0.0 0.0 0.0\n")
		sb.WriteString("0.0 0.0 0.0\n")
		sb.WriteString("0\n")
		sb.WriteString("0.0\n")

	} else if meta.ModelType == "TRAJECTORY" {
		sb.WriteString(fmt.Sprintf("%s\n", meta.OutputFile.Directory))
		sb.WriteString(fmt.Sprintf("%s\n", meta.OutputFile.FileName))
		sb.WriteString("0\n")
		sb.WriteString("100\n")
	}

	return sb.String(), nil
}