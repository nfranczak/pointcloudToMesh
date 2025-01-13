package convert

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"go.viam.com/rdk/logging"
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

func (g *gen) writeToFile(cloud pointcloud.PointCloud, pcdPath string) error {
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

func (g *gen) getMeshFromPointCloud() error {
	err := removeContents(g.workingDirectory + meshSubDir)
	if err != nil {
		return err
	}

	// generates a .ply file locally
	return g.meshSurfaceReconstruction()
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
		strconv.FormatFloat(g.radius, 'g', -1, 64),
		strconv.Itoa(g.maxNN),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func plyToBytes(path string) ([]byte, error) {
	return os.ReadFile(path)
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

func countFiles(path string) (int, error) {
	// Ensure the directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return 0, err
	}

	// Read the directory contents
	files, err := os.ReadDir(path)
	if err != nil {
		return 0, err
	}

	// Count the files
	return len(files), nil
}

func readFiles(path string) ([]pointcloud.PointCloud, error) {
	allClouds := []pointcloud.PointCloud{}
	// Ensure the directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	// Read the directory contents
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	// iterate through the files and convert them into a pointcloud object
	for _, f := range files {
		if strings.Contains(f.Name(), ".pcd") {
			fmt.Println("path + f.Name(): ", path+f.Name())
			pointCloudFile, err := os.Open(path + f.Name())
			if err != nil {
				fmt.Println("return err here 1")
				return nil, err
			}
			pc, err := pointcloud.ReadPCD(pointCloudFile)
			if err != nil {
				fmt.Println("return err here 2")
				return nil, err
			}
			allClouds = append(allClouds, pc)
		}

	}

	return allClouds, nil
}

func joinClouds(logger logging.Logger, allClouds []pointcloud.PointCloud) (pointcloud.PointCloud, error) {
	suplimentaryFuncs := []pointcloud.CloudAndOffsetFunc{}
	for _, cloud := range allClouds {
		suplimentaryFuncs = append(suplimentaryFuncs,
			func(context context.Context) (pointcloud.PointCloud, spatialmath.Pose, error) {
				return cloud, nil, nil
			},
		)
	}
	return pointcloud.MergePointClouds(context.Background(), suplimentaryFuncs, logger)
}
