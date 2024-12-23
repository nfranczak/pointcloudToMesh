package convert

import (
	"context"

	"github.com/golang/geo/r3"
	"github.com/mitchellh/mapstructure"
	"go.viam.com/rdk/components/arm"
	"go.viam.com/rdk/components/camera"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/generic"
	"go.viam.com/rdk/services/motion"
	"go.viam.com/rdk/spatialmath"
)

var Model = resource.NewModel("viam", "pcd-to-mesh", "converter")

func init() {
	resource.RegisterService(generic.API, Model, resource.Registration[resource.Resource, *Config]{Constructor: newConverter})
}

func newConverter(ctx context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (resource.Resource, error) {
	g := &gen{
		logger: logger,
	}

	if err := g.Reconfigure(ctx, deps, conf); err != nil {
		return nil, err
	}
	return g, nil
}

func (cfg *Config) Validate(path string) ([]string, error) {
	return []string{cfg.ArmName, cfg.CameraName, motion.Named("builtin").String()}, nil
}

type Config struct {
	ArmName    string `json:"arm_name"`
	CameraName string `json:"camera_name"`
}

type gen struct {
	resource.Resource
	resource.Named
	resource.TriviallyReconfigurable
	resource.TriviallyCloseable
	logger logging.Logger
	a      arm.Arm
	c      camera.Camera
	m      motion.Service
}

func (g *gen) Reconfigure(ctx context.Context, deps resource.Dependencies, conf resource.Config) error {
	config, err := resource.NativeConfig[*Config](conf)
	if err != nil {
		return err
	}

	a, err := arm.FromDependencies(deps, config.ArmName)
	if err != nil {
		return err
	}
	g.a = a

	c, err := camera.FromDependencies(deps, config.CameraName)
	if err != nil {
		return err
	}
	g.c = c

	m, err := motion.FromDependencies(deps, "builtin")
	if err != nil {
		return err
	}
	g.m = m

	g.logger.Info("done reconfiguring")
	return nil
}

func (g *gen) Name() resource.Name {
	return resource.NewName(generic.API, "pc-to-mesh")
}

func (g *gen) Close(ctx context.Context) error {
	return nil
}

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

// DoCommand echos input back to the caller.
func (g *gen) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	mesh, err := g.getMeshFromPC()
	if err != nil {
		return nil, err
	}
	g.logger.Infof("mesh.Triangles(): %v", mesh.Triangles())
	g.logger.Infof("len(mesh.Triangles()): %v", len(mesh.Triangles()))
	var asMap []map[string]any

	// v := convertTriangle(mesh)
	convertible := make([]MyTriangle, 0, len(mesh.Triangles()))
	for _, t := range mesh.Triangles() {
		convertible = append(convertible, convertTriangle(t))
	}
	mapstructure.Decode(convertible, &asMap)
	println("decoded to", len(asMap))

	return map[string]interface{}{"mesh_triangles": asMap}, nil
}
