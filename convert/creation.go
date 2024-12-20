package convert

import (
	"bufio"
	"context"
	"errors"
	"os"
	"os/exec"

	"github.com/chenzhekl/goply"

	"github.com/golang/geo/r3"
	"go.viam.com/rdk/pointcloud"
	"go.viam.com/rdk/spatialmath"
)

func (g *gen) getMeshFromPC() (*spatialmath.Mesh, error) {
	ctx := context.Background()

	// Get the camera from the robot
	realsense := g.c

	myPointcloud, err := realsense.NextPointCloud(ctx)
	if err != nil {
		return nil, err
	}

	err = writePCD("/Users/nick/Desktop/whiteboard.pcd", myPointcloud)
	if err != nil {
		return nil, err
	}

	err = runPoissonReconstruction()
	if err != nil {
		return nil, err
	}
	// above generates a .ply file locally

	mesh, err := ReadPLY("some path here")
	if err != nil {
		return nil, err
	}

	return mesh, nil
}

func writePCD(filepath string, whiteBoardPointCloud pointcloud.PointCloud) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = pointcloud.ToPCD(whiteBoardPointCloud, file, pointcloud.PCDAscii) // or pointcloud.BINARY depending on your needs
	if err != nil {
		return err
	}
	return nil
}

// Call Open3D to perform Poisson Surface Reconstruction
func runPoissonReconstruction() error {
	// Ensure Open3D is installed and callable as a Python script
	// 	cmd := exec.Command("python3.10 draw.py", "-c", fmt.Sprintf(`
	// import open3d as o3d
	// pcd = o3d.io.read_point_cloud("%s")
	// mesh, _ = o3d.geometry.TriangleMesh.create_from_point_cloud_poisson(pcd, depth=8)
	// o3d.io.write_triangle_mesh("%s", mesh)
	// `, inputFile, outputFile))

	// ABOVE IS THE OLD VERSION FROM WHEN I WAS FIRST EXPERIMENTING WITH EXEC

	cmd := exec.Command("python3.10 draw.py")
	// not sure if the above is done properly

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ReadPLY(path string) (*spatialmath.Mesh, error) {
	readerRaw, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(readerRaw)
	ply := goply.New(reader)
	vertices := ply.Elements("vertex")
	faces := ply.Elements("face")
	triangles := []*spatialmath.Triangle{}
	for _, face := range faces {

		pts := []r3.Vector{}
		idxIface := face["vertex_indices"]
		for _, i := range idxIface.([]interface{}) {
			pts = append(pts, r3.Vector{
				X: 1000 * vertices[int(i.(uint32))]["x"].(float64),
				Y: 1000 * vertices[int(i.(uint32))]["y"].(float64),
				Z: 1000 * vertices[int(i.(uint32))]["z"].(float64)})
		}
		if len(pts) != 3 {
			return nil, errors.New("triangle did not have three points")
		}
		tri := spatialmath.NewTriangle(pts[0], pts[1], pts[2])
		triangles = append(triangles, tri)
	}
	return spatialmath.NewMesh(spatialmath.NewZeroPose(), triangles), nil
}
