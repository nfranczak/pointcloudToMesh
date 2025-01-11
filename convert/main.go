package convert

import (
	"context"
	"errors"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/generic"
)

var Model = resource.NewModel("viam", "pcd-to-mesh", "converter")

const (
	fileName   = "merged.pcd"
	meshSubDir = "mesh/"
	lodPLY     = "lod_100.ply"
)

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
	return []string{
		// cfg.WorkingDirectory,
		// cfg.MeshAlgorithm,
		// cfg.PythonPath,
		// cfg.DownSample,
		// motion.Named("builtin").String(),
	}, nil
}

type Config struct {
	WorkingDirectory string `json:"working_directory"`
	MeshAlgorithm    string `json:"mesh_algorithm"`
	PythonPath       string `json:"python_path"`
	// DownSample       float64 `json:"down_sample`
}

type gen struct {
	resource.Resource
	resource.Named
	resource.TriviallyReconfigurable
	resource.TriviallyCloseable
	logger           logging.Logger
	workingDirectory string
	pythonPath       string
	meshAlgorithm    string
}

func (g *gen) Reconfigure(ctx context.Context, deps resource.Dependencies, conf resource.Config) error {
	config, err := resource.NativeConfig[*Config](conf)
	if err != nil {
		return err
	}
	g.workingDirectory = config.WorkingDirectory
	g.pythonPath = config.PythonPath

	// make sure we specified a valid mesh reconstruction algorithm name
	switch algorithmName := config.MeshAlgorithm; algorithmName {
	case "bpa":
	case "poisson":
	case "cube":
	default:
		return errors.New("did not specify valid mesh algorithm")
	}
	g.meshAlgorithm = config.MeshAlgorithm

	g.logger.Info("done reconfiguring")
	return nil
}

func (g *gen) Name() resource.Name {
	return resource.NewName(generic.API, "pc-to-mesh")
}

func (g *gen) Close(ctx context.Context) error {
	// todo: implement
	return nil
}

// DoCommand echos input back to the caller.
func (g *gen) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	pcInterface, ok := cmd["pointcloud"]
	if !ok {
		return nil, errors.New("command was incorrectly specified")
	}
	pc, err := getPointCloudFromBytes(pcInterface)
	if err != nil {
		return nil, err
	}
	g.logger.Infof("got the pointcloud from bytes: %v", pc)

	// write to a .pcd file so that mesh reconstruction can read from it
	err = g.writeToFile(pc)
	if err != nil {
		return nil, err
	}
	g.logger.Infof("wrote the the pointcloud to a file")

	// get the spatialmath.Mesh
	mesh, err := g.getMeshFromPointCloud()
	if err != nil {
		return nil, err
	}

	// generate json representation of mesh
	return map[string]interface{}{"mesh_triangles": generateMeshJson(mesh)}, nil
}
