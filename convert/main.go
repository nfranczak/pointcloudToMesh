package convert

import (
	"context"
	"sync"

	"go.viam.com/rdk/components/arm"
	"go.viam.com/rdk/components/camera"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/robot/client"
	"go.viam.com/rdk/services/generic"
	"go.viam.com/rdk/services/motion"
	"go.viam.com/rdk/services/vision"
	"go.viam.com/rdk/spatialmath"
	rutils "go.viam.com/rdk/utils"
	"go.viam.com/utils"
	"go.viam.com/utils/rpc"
)

var Model = resource.NewModel("viam", "pcd-to-mesh", "converter")

func init() {
	resource.RegisterService(generic.API, Model, resource.Registration[resource.Resource, *Config]{Constructor: newConverter})
}

func newConverter(ctx context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (resource.Resource, error) {
	address, err := rutils.AssertType[string](conf.Attributes["address"])
	if err != nil {
		return nil, err
	}
	entity, err := rutils.AssertType[string](conf.Attributes["entity"])
	if err != nil {
		return nil, err
	}
	payload, err := rutils.AssertType[string](conf.Attributes["payload"])
	if err != nil {
		return nil, err
	}

	g := &gen{
		logger:  logger,
		address: address,
		entity:  entity,
		payload: payload,
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
	Address    string `json:"address"`
	Entity     string `json:"entity"`
	Payload    string `json:"payload"`
}

type gen struct {
	mu sync.Mutex
	resource.Resource
	resource.Named
	resource.TriviallyReconfigurable
	resource.TriviallyCloseable
	logger                                                   logging.Logger
	address, entity, payload                                 string
	robotClient                                              *client.RobotClient
	a                                                        arm.Arm
	c                                                        camera.Camera
	s                                                        sensor.Sensor
	m                                                        motion.Service
	v                                                        vision.Service
	deltaXPos, deltaYPos, deltaXNeg, deltaYNeg, bottleHeight float64
	status                                                   string
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

	utils.PanicCapturingGo(g.getRobotClient)

	g.logger.Info("done reconfiguring")
	return nil
}

func (g *gen) getRobotClient() {
	machine, err := client.New(
		context.Background(),
		g.address,
		g.logger,
		client.WithDialOptions(rpc.WithEntityCredentials(
			g.entity,
			rpc.Credentials{
				Type:    rpc.CredentialsTypeAPIKey,
				Payload: g.payload,
			})),
	)
	if err != nil {
		g.logger.Fatal(err)
	}
	g.robotClient = machine
}

func (g *gen) Name() resource.Name {
	return resource.NewName(generic.API, "pc-to-mesh")
}

func (g *gen) Close(ctx context.Context) error {
	return g.robotClient.Close(ctx)
}

// DoCommand echos input back to the caller.
func (g *gen) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	mesh, err := g.getMeshFromPC()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"mesh ": spatialmath.NewGeometriesToProto([]spatialmath.Geometry{mesh})}, nil
}
