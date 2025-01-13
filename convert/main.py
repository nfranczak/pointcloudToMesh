import argparse
from helpers import get_point_cloud, de_duplicate
from algos import create_BPA_mesh, create_poisson_mesh


def main():
    print("WE ARE INSIDE THE PYTHON FILE")
    # Create the argument parser
    parser = argparse.ArgumentParser(description="Pass arguments to a function.")
    
    # Define the command-line arguments
    parser.add_argument('input_path', type=str, help="First argument -- input path")
    parser.add_argument('output_path', type=str, help="Second argument -- output path")
    parser.add_argument('data_name', type=str, help="Third argument -- data_name")
    parser.add_argument('algo_name', type=str, help="Third argument -- algo_name")
    
    # Parse the arguments
    args = parser.parse_args()
    print("args.input_path: ", args.input_path)
    print("type(args.input_path): ", type(args.input_path))
    print("args.output_path: ", args.output_path)
    print("type(args.output_path): ", type(args.output_path))
    print("args.data_name: ", args.data_name)
    print("type(args.data_name): ", type(args.data_name))
    print("args.algo_name: ", args.algo_name)
    print("type(args.algo_name): ", type(args.algo_name))
    
    input_path = args.input_path
    output_path = args.output_path
    data_name = args.data_name
    algo_name = args.algo_name
    
    print("input_path: ", input_path)
    print("output_path: ", output_path)
    print("data_name: ", data_name)
    print("algo_name: ", algo_name)
    
    
    pcd = get_point_cloud(input_path, data_name)
    
    # Downsample pointcloud
    downSampled_PointCloud = de_duplicate(pcd)
    
    # Outlier removal
    # ToDo
    
    # Create the actual mesh
    if algo_name == "bpa":
        print("going into bpa algo")
        create_BPA_mesh(downSampled_PointCloud, output_path)
    elif algo_name == "poisson":
        print("going into poisson algo")
        create_poisson_mesh((downSampled_PointCloud, output_path))


# Entry point of the script
if __name__ == "__main__":
    main()
