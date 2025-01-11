import numpy as np
import open3d as o3d

def get_point_cloud(input_path, dataname):
    # Load the point cloud from the .pcd file
    point_cloud = o3d.io.read_point_cloud(input_path+dataname)

    # Extract points as a NumPy array
    points = np.asarray(point_cloud.points)

    # If needed, perform slicing or other operations on the NumPy array
    points_subset = points[:, :3] 

    # Convert to Open3D PointCloud
    pcd = o3d.geometry.PointCloud()
    pcd.points = o3d.utility.Vector3dVector(points_subset)

    # Estimate normals for pcd directly
    pcd.estimate_normals(search_param=o3d.geometry.KDTreeSearchParamHybrid(radius=0.1, max_nn=30))

    # Check if normals are assigned
    if not pcd.has_normals():
        raise ValueError("Normals could not be computed for the point cloud.")
    
    return pcd

def lod_mesh_export(mesh, lods, extension, path):
    mesh_lods = {}
    for i in lods:
        mesh_lod = mesh.simplify_quadric_decimation(i)
        # Save the mesh in ASCII format
        o3d.io.write_triangle_mesh(
            path + f"lod_{i}{extension}", 
            mesh_lod, 
            write_ascii=True
        )
        mesh_lods[i] = mesh_lod
    print(f"Generation of {len(lods)} LoD successful")
    return mesh_lods

def de_duplicate(pcd):
    return pcd.voxel_down_sample(voxel_size=0.01)