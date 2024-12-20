import numpy as np
import open3d as o3d

# this one writes the .ply in ascii whcih is what we need
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


input_path = "/Users/nick/Desktop/"
output_path = "/Users/nick/Desktop/mesh/"
# dataname = "whiteboard.xyz"
dataname = "whiteboard.pcd"

# Load the point cloud from the .xyz file
point_cloud = o3d.io.read_point_cloud(input_path+dataname)

# Extract points as a NumPy array
points = np.asarray(point_cloud.points)

# If needed, perform slicing or other operations on the NumPy array
points_subset = points[:, :3]  # Example of slicing if needed

# Convert to Open3D PointCloud
pcd = o3d.geometry.PointCloud()
pcd.points = o3d.utility.Vector3dVector(points_subset)

# Estimate normals for pcd directly
pcd.estimate_normals(search_param=o3d.geometry.KDTreeSearchParamHybrid(radius=0.1, max_nn=30))

# Check if normals are assigned
if not pcd.has_normals():
    raise ValueError("Normals could not be computed for the point cloud.")

# # poisson surface reconstruction method
poisson_mesh = o3d.geometry.TriangleMesh.create_from_point_cloud_poisson(pcd, depth=8, width=0, scale=1.1, linear_fit=False)[0]
poisson_mesh.remove_degenerate_triangles()
poisson_mesh.remove_duplicated_triangles()
poisson_mesh.remove_duplicated_vertices()
poisson_mesh.remove_non_manifold_edges()

# triangles = np.asarray(poisson_mesh.triangles)

# print("Triangles shape:", triangles.shape)
# print(triangles)
bbox = pcd.get_axis_aligned_bounding_box()
p_mesh_crop = poisson_mesh.crop(bbox)



o3d.io.write_triangle_mesh(output_path+"p_mesh_c.ply", p_mesh_crop)

# Generate LoDs
# IMPORTANT THE NUMBER SLICE THAT IS PASSED IN MEANS THIS
# NUM == NUM OF TRINAGLES THAT COMPRISE THE MESH
lod_mesh_export(p_mesh_crop, [100000, 50000, 10000, 1000, 100], ".ply", output_path)
# my_lods = lod_mesh_export(p_mesh_crop, [100000, 50000, 10000, 1000, 100], ".ply", output_path)

# lod_mesh = my_lods[100]
# print(lod_mesh)



# # # Visualize one LoD
# o3d.visualization.draw_geometries([lod_mesh])
