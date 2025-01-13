package convert

import (
	"context"
	"errors"
	"strconv"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/generic"
)

var Model = resource.NewModel("viam", "pcd-to-mesh", "converter")

const (
	fileName         = "merged.pcd"
	meshSubDir       = "mesh/"
	pointcloudSubDir = "modulePointClouds/"
	lodPLY           = "lod_100.ply"
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
	WorkingDirectory string  `json:"working_directory"`
	MeshAlgorithm    string  `json:"mesh_algorithm"`
	PythonPath       string  `json:"python_path"`
	Radius           float64 `json:"radius"`
	MaxNN            int     `json:"max_nn"`
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
	radius           float64
	maxNN            int
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
	// reads in a passed in cloud and returns it as a .ply file representing a mesh
	if pcInterface, ok := cmd["pointcloud"]; ok {
		pc, err := getPointCloudFromBytes(pcInterface)
		if err != nil {
			return nil, err
		}
		g.logger.Infof("got the pointcloud from bytes: %v", pc)

		// write to a .pcd file so that mesh reconstruction can read from it
		err = g.writeToFile(pc, g.workingDirectory+fileName)
		if err != nil {
			return nil, err
		}
		g.logger.Infof("wrote the the pointcloud to a file")

		// get the mesh as a .ply file locally
		err = g.getMeshFromPointCloud()
		if err != nil {
			return nil, err
		}

		// convert ply file into a slice of bytes which are then sent over the wire
		plyFileAsBytes, err := plyToBytes(g.workingDirectory + meshSubDir + lodPLY)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"plyFileBytes": plyFileAsBytes}, nil
	}

	// adds clouds to storage
	if pcInterface, ok := cmd["addCloud"]; ok {
		pc, err := getPointCloudFromBytes(pcInterface)
		if err != nil {
			return nil, err
		}
		g.logger.Infof("got the pointcloud from bytes: %v", pc)

		// check how many files exist in
		pointCloudStoragePath := g.workingDirectory + pointcloudSubDir
		g.logger.Infof("counting the files in this path: %s", pointCloudStoragePath)
		numFiles, err := countFiles(pointCloudStoragePath)
		if err != nil {
			return nil, err
		}

		// write the pointcloud to file
		writePointCloudPath := pointCloudStoragePath + "cloud" + strconv.Itoa(numFiles) + ".pcd"
		g.logger.Infof("writing pointcloud here: %s", writePointCloudPath)
		err = g.writeToFile(pc, writePointCloudPath)
		if err != nil {
			return nil, err
		}
		g.logger.Infof("wrote the the pointcloud to a file")

		return nil, nil
	}

	// reads the clouds in storage, merges them together and returns the representing mesh
	if _, ok := cmd["merge"]; ok {
		// read the pointclouds from path
		pointCloudStoragePath := g.workingDirectory + pointcloudSubDir
		g.logger.Infof("reading the files in this path: %s", pointCloudStoragePath)
		allClouds, err := readFiles(pointCloudStoragePath)
		if err != nil {
			return nil, err
		}

		// merge them together
		mergedCloud, err := joinClouds(g.logger, allClouds)
		if err != nil {
			return nil, err
		}

		// write the merged cloud to a file
		err = g.writeToFile(mergedCloud, g.workingDirectory+fileName)
		if err != nil {
			return nil, err
		}
		g.logger.Infof("wrote the the pointcloud to a file")

		// get the spatialmath.Mesh
		err = g.getMeshFromPointCloud()
		if err != nil {
			return nil, err
		}

		// wipe the pointcloud storage clean
		defer removeContents(pointCloudStoragePath)

		plyFileAsBytes, err := plyToBytes(g.workingDirectory + meshSubDir + lodPLY)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{"plyFileBytes": plyFileAsBytes}, nil
	}

	return cmd, nil
}
