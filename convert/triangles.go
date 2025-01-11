package convert

import (
	"github.com/golang/geo/r3"
	"github.com/mitchellh/mapstructure"
	"go.viam.com/rdk/spatialmath"
)

type MyVec struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type MyTriangle struct {
	P0     MyVec `json:"p0"`
	P1     MyVec `json:"p1"`
	P2     MyVec `json:"p2"`
	Normal MyVec `json:"normal"`
}

func convertPoint(vec r3.Vector) MyVec {
	return MyVec{vec.X, vec.Y, vec.Z}
}

func convertTriangle(t *spatialmath.Triangle) MyTriangle {
	points := t.Points()
	return MyTriangle{
		P0:     convertPoint(points[0]),
		P1:     convertPoint(points[1]),
		P2:     convertPoint(points[2]),
		Normal: convertPoint(t.Normal()),
	}
}

func generateMeshJson(mesh *spatialmath.Mesh) []map[string]any {
	var asMap []map[string]any

	convertible := make([]MyTriangle, 0, len(mesh.Triangles()))
	for _, t := range mesh.Triangles() {
		convertible = append(convertible, convertTriangle(t))
	}
	mapstructure.Decode(convertible, &asMap)
	return asMap
}
