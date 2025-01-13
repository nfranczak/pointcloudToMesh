package convert

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/chenzhekl/goply"

	"github.com/golang/geo/r3"
	"go.viam.com/rdk/pointcloud"
	"go.viam.com/rdk/spatialmath"
	"go.viam.com/rdk/utils"
)

func getPointCloudFromBytes(pcInterface interface{}) (pointcloud.PointCloud, error) {
	// get the pointcloud bytes from Extras and convert back into a pointcloud
	pcdSlice, err := utils.AssertType[[]interface{}](pcInterface)
	if err != nil {
		return nil, err
	}
	uInt8Slice := []uint8{}
	for _, v := range pcdSlice {
		data, err := utils.AssertType[float64](v)
		if err != nil {
			return nil, err
		}
		uInt8Slice = append(uInt8Slice, uint8(data))
	}

	data, err := utils.AssertType[[]byte](uInt8Slice)
	if err != nil {
		return nil, err
	}

	pc, err := pointcloud.ReadPCD(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return pc, nil
}

func (g *gen) writeToFile(cloud pointcloud.PointCloud) error {
	pcdPath := g.workingDirectory + fileName
	// pcdPath should  be "/Users/nick/Desktop/whiteboard.pcd"
	g.logger.Infof("we are writing the pointcloud to this file: %s", pcdPath)
	file, err := os.Create(pcdPath)
	if err != nil {
		return err
	}
	err = pointcloud.ToPCD(cloud, file, pointcloud.PCDAscii)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}

func (g *gen) getMeshFromPointCloud() (*spatialmath.Mesh, error) {
	err := removeContents(g.workingDirectory + meshSubDir)
	if err != nil {
		return nil, err
	}

	// generates a .ply file locally
	err = g.meshSurfaceReconstruction()
	if err != nil {
		return nil, err
	}

	g.logger.Infof("removed the ")

	mesh, err := readPLY(g.workingDirectory + meshSubDir + lodPLY)
	if err != nil {
		return nil, err
	}

	return mesh, nil
}

func removeContents(dirPath string) error {
	// List all the files and subdirectories in the specified directory
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("unable to read directory: %v", err)
	}

	// Iterate through all the files/subdirectories and remove them
	for _, file := range files {
		filePath := filepath.Join(dirPath, file.Name())

		// If it's a directory, recursively remove its contents
		if file.IsDir() {
			err = os.RemoveAll(filePath)
			if err != nil {
				return fmt.Errorf("unable to remove directory %s: %v", filePath, err)
			}
		} else {
			// If it's a file, remove it
			err = os.Remove(filePath)
			if err != nil {
				return fmt.Errorf("unable to remove file %s: %v", filePath, err)
			}
		}
	}

	return nil
}

func (g *gen) meshSurfaceReconstruction() error {
	pathToFile := "../../convert/main.py"

	g.logger.Infof("g.pythonPath %s", g.pythonPath)
	g.logger.Infof("pathToFile: %s", pathToFile)
	g.logger.Infof("g.workingDirectory: %s", g.workingDirectory)
	g.logger.Infof("g.workingDirectory+meshSubDir: %s", g.workingDirectory+meshSubDir)
	g.logger.Infof("fileName: %s", fileName)
	g.logger.Infof("g.meshAlgorithm: %s", g.meshAlgorithm)

	cmd := exec.Command(
		g.pythonPath,
		pathToFile,
		g.workingDirectory,
		g.workingDirectory+meshSubDir,
		fileName,
		g.meshAlgorithm,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func readPLY(path string) (*spatialmath.Mesh, error) {
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
