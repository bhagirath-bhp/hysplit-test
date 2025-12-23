package main

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fogleman/gg"
)

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
	T      int64              `json:"t"`
	Bbox   map[string]float64 `json:"bbox"`
	Base64 string             `json:"base64"`
}

func lonToX(lon float64) float64 {
	return (lon + 180.0) * (256.0 / 360.0)
}

func latToY(lat float64) float64 {
	latRad := lat * math.Pi / 180.0
	return (128.0 / math.Pi) * (math.Pi - math.Log(math.Tan((math.Pi/4.0)+(latRad/2.0))))
}

func parseKmlColor(kmlColor string) color.RGBA {
	if len(kmlColor) != 8 {
		fmt.Println("Invalid KML color format:", kmlColor)
		return color.RGBA{0, 0, 0, 128}
	}
	// KML is AABBGGRR
	a, _ := strconv.ParseUint(kmlColor[0:2], 16, 8)
	b, _ := strconv.ParseUint(kmlColor[2:4], 16, 8)
	g, _ := strconv.ParseUint(kmlColor[4:6], 16, 8)
	r, _ := strconv.ParseUint(kmlColor[6:8], 16, 8)
	return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
}

func parseCoordinates(s string) [][]float64 {
	var points [][]float64
	lines := strings.Fields(s)
	for _, line := range lines {
		parts := strings.Split(line, ",")
		if len(parts) >= 2 {
			lon, _ := strconv.ParseFloat(parts[0], 64)
			lat, _ := strconv.ParseFloat(parts[1], 64)
			points = append(points, []float64{lon, lat})
		}
	}
	return points
}

func extractTimestamp(f Folder) int64 {
	if f.TimeSpan.Begin != "" {
		formats := []string{time.RFC3339, "2006-01-02T15:04:05Z", "2006-01-02T15:04:05"}
		for _, fmtStr := range formats {
			t, err := time.Parse(fmtStr, f.TimeSpan.Begin)
			if err == nil {
				return t.Unix()
			}
		}
	}
	re := regexp.MustCompile(`Valid:(\d{8})\s+(\d{4})`)
	match := re.FindStringSubmatch(f.Name)
	if len(match) == 3 {
		t, err := time.Parse("20060102 1504", match[1]+" "+match[2])
		if err == nil {
			return t.Unix()
		}
	}
	return time.Now().Unix()
}

func validateKmlContent(s string) (bool, error) {
	s = strings.TrimSpace(s)
	if !strings.HasSuffix(s, "</kml>") {
		return false, fmt.Errorf("KML file is incomplete or malformed")
	}
	return true, nil
}

func ProcessKml(filePath string) ([]KmlResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	sContent := strings.TrimSpace(string(content))
	isValid, err := validateKmlContent(sContent)
	if !isValid {
		return nil, err
	}

	var root KmlRoot
	if err := xml.Unmarshal([]byte(sContent), &root); err != nil {
		return nil, fmt.Errorf("XML unmarshal error: %v", err)
	}
	styles := make(map[string]color.RGBA)
	for _, s := range root.Document.Styles {
		styles["#"+s.ID] = parseKmlColor(s.PolyStyle.Color)
		fmt.Println("Parsed style ID:", s.ID, "for color:", s.PolyStyle.Color, "with RGBA:", styles["#"+s.ID])
	}

	var results []KmlResult

	for _, folder := range root.Document.Folders {
		if !strings.Contains(folder.Name, "Concentration") {
			continue
		}

		var folderPoints [][]float64
		for _, pm := range folder.Placemarks {
			for _, poly := range pm.MultiGeometry.Polygons {
				folderPoints = append(folderPoints, parseCoordinates(poly.OuterBoundary)...)
			}
		}

		if len(folderPoints) == 0 {
			continue
		}
		minLon, minLat, maxLon, maxLat := 180.0, 90.0, -180.0, -90.0
		for _, p := range folderPoints {
			if p[0] < minLon {
				minLon = p[0]
			}
			if p[0] > maxLon {
				maxLon = p[0]
			}
			if p[1] < minLat {
				minLat = p[1]
			}
			if p[1] > maxLat {
				maxLat = p[1]
			}
		}
		minX, maxX := lonToX(minLon), lonToX(maxLon)
		minY, maxY := latToY(maxLat), latToY(minLat)
		const imgSize = 1024
		dc := gg.NewContext(imgSize, imgSize)
		for _, pm := range folder.Placemarks {
			c, ok := styles[pm.StyleUrl]
			fmt.Println("Using style URL:", pm.StyleUrl, "Color found:", ok)
			if !ok {
				c = color.RGBA{255, 0, 0, 128}
			}
			dc.SetColor(c)
			fmt.Println("Drawing Placemark:", pm.Name, "with color:", c)

			for _, poly := range pm.MultiGeometry.Polygons {
				pts := parseCoordinates(poly.OuterBoundary)
				for i, p := range pts {
					x := (lonToX(p[0]) - minX) / (maxX - minX) * imgSize
					y := (latToY(p[1]) - minY) / (maxY - minY) * imgSize
					if i == 0 {
						dc.MoveTo(x, y)
					} else {
						dc.LineTo(x, y)
					}
				}
				dc.Fill()
			}
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, dc.Image()); err != nil {
			continue
		}
		b64 := base64.StdEncoding.EncodeToString(buf.Bytes())

		results = append(results, KmlResult{
			T: extractTimestamp(folder),
			Bbox: map[string]float64{
				"west": minLon, "south": minLat, "east": maxLon, "north": maxLat,
			},
			Base64: "data:image/png;base64," + b64,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].T < results[j].T
	})

	return results, nil
}

func main() {
	logFile, err := os.OpenFile("output.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	kmlPath := "/workspaces/hysplit-test/programfiles/hysplit/working/HYSPLIT_ps.kml"
	results, err := ProcessKml(kmlPath)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	message := fmt.Sprintf("Successfully processed %d time segments.\n", len(results))
	log.Print(message)
	fmt.Print(message)

	for _, res := range results {
		line := fmt.Sprintf("Time: %d, BBox: %+v, Base64 Length: %d\n", res.T, res.Bbox, res.Base64)
		log.Print(line)
		// fmt.Print(line)
	}
}
