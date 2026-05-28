package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/WrenchRobotics/meshcat-go/geometry"
	"github.com/WrenchRobotics/meshcat-go/meshcat"
)

func main() {
	// Create a new Meshcat viewer window
	window, err := meshcat.NewViewerWindow()
	if err != nil {
		log.Fatal(err)
	}
	defer window.Close()

	// Open the viewer and wait for the browser websocket to connect before
	// issuing the initial scene commands.
	if err := window.Open(); err != nil {
		log.Fatal(err)
	}

	// Get the Visualizer from the window
	v := window.Visualizer()
	if err := v.At("Background").SetProperty("top_color", []float64{1, 0, 0}); err != nil {
		log.Fatal(err)
	}
	cube := v.At("cube")

	// Set a red box mesh at /cube so the animation transforms the cube
	// instead of the scene root.
	mat := geometry.NewMeshBasicMaterial()
	mat.Color = 0xff0000
	box := geometry.NewMesh(geometry.NewBox([3]float64{1, 1, 1}), mat)
	if err := cube.SetObject(box); err != nil {
		log.Fatal(err)
	}

	drawTimes := make([]time.Duration, 0, 200)
	for i := 0; i < 200; i++ {
		theta := float64(i+1) / 100.0 * 2.0 * math.Pi
		started := time.Now()

		if err := cube.SetTransform(rotationZ(theta)); err != nil {
			log.Fatal(err)
		}

		drawTimes = append(drawTimes, time.Since(started))
		time.Sleep(10 * time.Millisecond)
	}

	var total time.Duration
	for _, drawTime := range drawTimes {
		total += drawTime
	}

	fmt.Println(total.Seconds() / float64(len(drawTimes)))
}

func rotationZ(theta float64) [4][4]float64 {
	cosTheta := math.Cos(theta)
	sinTheta := math.Sin(theta)

	return [4][4]float64{
		{cosTheta, -sinTheta, 0, 0},
		{sinTheta, cosTheta, 0, 0},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
	}
}
