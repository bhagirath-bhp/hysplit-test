package main

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"image/color"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fogleman/gg"
)

// --- Types for KML Parsing ---

type KmlRoot struct {
	XMLName  xml.Name `xml:"kml"`
	Document Document `xml:"Document"`
}

type Document struct {
	Styles    []Style    `xml:"Style"`
	StyleMaps []StyleMap `xml:"StyleMap"`
	Folders   []Folder   `xml:"Folder"`
}

type Style struct {
	ID        string    `xml:"id,attr"`
	PolyStyle PolyStyle `xml:"PolyStyle"`
}

type PolyStyle struct {
	Color string `xml:"color"`
}

type StyleMap struct {
	ID string `xml:"id,attr"`
}

type Folder struct {
	Name       string      `xml:"name"`
	TimeSpan   TimeSpan    `xml:"TimeSpan"`
	Placemarks []Placemark `xml:"Placemark"`
}

type TimeSpan struct {
	Begin string `xml:"begin"`
}

type Placemark struct {
	Name          string        `xml:"name"`
	StyleUrl      string        `xml:"styleUrl"`
	MultiGeometry MultiGeometry `xml:"MultiGeometry"`
}

type MultiGeometry struct {
	Polygons []Polygon `xml:"Polygon"`
}

type Polygon struct {
	OuterBoundary string `xml:"outerBoundaryIs>LinearRing>coordinates"`
}

type KmlResult struct {
	T      int64             `json:"t"`
	Bbox   map[string]float64 `json:"bbox"`
	Base64 string            `json:"base64"`
}

// --- Mercator & Coordinate Helpers ---

func lonToX(lon float64) float64 {
	return (lon + 180) * (256 / 360.0)
}

func latToY(lat float64) float64 {
	latRad := lat * math.Pi / 180.0
	return (128 / math.Pi) * (math.Pi - math.Log(math.Tan((math.Pi/4.0)+(latRad/2.0))))
}

func parseColor(kmlColor string) color.RGBA {
	// KML is AABBGGRR
	if len(kmlColor) != 8 {
		return color.RGBA{0, 0, 0, 255}
	}
	a, _ := strconv.ParseUint(kmlColor[0:2], 16, 8)
	b, _ := strconv.ParseUint(kmlColor[2:4], 16, 8)
	g, _ := strconv.ParseUint(kmlColor[4:6], 16, 8)
	r, _ := strconv.ParseUint(kmlColor[6:8], 16, 8)
	return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
}

// --- Main Processing Function ---

func ExtractKmlSegments(filePath string) ([]KmlResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var root KmlRoot
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, err
	}

	styleMap := make(map[string]color.RGBA)
	for _, s := range root.Document.Styles {
		styleMap["#"+s.ID] = parseColor(s.PolyStyle.Color)
	}

	var finalResults []KmlResult

	for _, folder := range root.Document.Folders {
		if !strings.Contains(folder.Name, "Concentration") {
			continue
		}

		// 1. Extract Timestamp
		timestamp := extractTimestamp(folder)
		
		// 2. Collect all points for BBox and calculate Mercator bounds
		var allCoords [][]float64
		for _, pm := range folder.Placemarks {
			for _, poly := range pm.MultiGeometry.Polygons {
				points := parseKmlCoords(poly.OuterBoundary)
				allCoords = append(allCoords, points...)
			}
		}

		if len(allCoords) == 0 {
			continue
		}

		minLon, minLat, maxLon, maxLat := 180.0, 90.0, -180.0, -90.0
		for _, pt := range allCoords {
			if pt[0] < minLon { minLon = pt[0] }
			if pt[0] > maxLon { maxLon = pt[0] }
			if pt[1] < minLat { minLat = pt[1] }
			if pt[1] > maxLat { maxLat = pt[1] }
		}

		// Convert bounds to Mercator space
		minX, maxX := lonToX(minLon), lonToX(maxLon)
		minY, maxY := latToY(maxLat), latToY(minLat) // Y is inverted in Mercator

		// 3. Render
		const size = 1024
		dc := gg.NewContext(size, size)
		
		for _, pm := range folder.Placemarks {
			fillColor, exists := styleMap[pm.StyleUrl]
			if !exists {
				fillColor = color.RGBA{255, 0, 0, 128}
			}
			dc.SetColor(fillColor)

			for _, poly := range pm.MultiGeometry.Polygons {
				points := parseKmlCoords(poly.OuterBoundary)
				for i, pt := range points {
					// Map Mercator X/Y to image pixel space
					x := (lonToX(pt[0]) - minX) / (maxX - minX) * size
					y := (latToY(pt[1]) - minY) / (maxY - minY) * size
					if i == 0 {
						dc.MoveTo(x, y)
					} else {
						dc.LineTo(x, y)
					}
				}
				dc.FillPreserve()
				dc.SetRGB(0, 0, 0)
				dc.SetLineWidth(1)
				dc.Stroke()
			}
		}

		// 4. Encode to Base64
		imgBase64, _ := dc.ToPNG() // Note: This is an internal helper in some gg versions, or use Encode
		b64Str := base64.StdEncoding.EncodeToString(imgBase64)

		finalResults = append(finalResults, KmlResult{
			T: timestamp,
			Bbox: map[string]float64{
				"west": minLon, "south": minLat, "east": maxLon, "north": maxLat,
			},
			Base64: "data:image/png;base64," + b64Str,
		})
	}

	sort.Slice(finalResults, func(i, j int) bool {
		return finalResults[i].T < finalResults[j].T
	})

	return finalResults, nil
}

func parseKmlCoords(s string) [][]float64 {
	var res [][]float64
	fields := strings.Fields(s)
	for _, f := range fields {
		parts := strings.Split(f, ",")
		if len(parts) >= 2 {
			lon, _ := strconv.ParseFloat(parts[0], 64)
			lat, _ := strconv.ParseFloat(parts[1], 64)
			res = append(res, []float64{lon, lat})
		}
	}
	return res
}

func extractTimestamp(f Folder) int64 {
	// Try TimeSpan Begin (ISO8601)
	if f.TimeSpan.Begin != "" {
		t, err := time.Parse(time.RFC3339, f.TimeSpan.Begin)
		if err == nil { return t.Unix() }
	}
	// Fallback to Regex on Name
	re := regexp.MustCompile(`Valid:(\d{8})\s+(\d{4})`)
	match := re.FindStringSubmatch(f.Name)
	if len(match) == 3 {
		t, err := time.Parse("20060102 1504", match[1]+" "+match[2])
		if err == nil { return t.Unix() }
	}
	return time.Now().Unix()
}