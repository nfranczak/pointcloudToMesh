import numpy as np
import open3d as o3d
from helpers import lod_mesh_export

def create_BPA_mesh(pcd, output_path):
    # Ball Pivoting Algorithm for surface reconstruction
    distances = pcd.compute_nearest_neighbor_distance()
    avg_dist = np.mean(distances)
    radius = 3 * avg_dist

    # Create mesh using Ball Pivoting
    bpa_mesh = o3d.geometry.TriangleMesh.create_from_point_cloud_ball_pivoting(
        pcd, o3d.utility.DoubleVector([radius, radius * 2])
    )

    # Simplify and clean the mesh
    dec_mesh = bpa_mesh.simplify_quadric_decimation(100000)
    dec_mesh.remove_degenerate_triangles()
    dec_mesh.remove_duplicated_triangles()
    dec_mesh.remove_duplicated_vertices()
    dec_mesh.remove_non_manifold_edges()

    # Save the mesh
    o3d.io.write_triangle_mesh(output_path + "bpa_mesh.ply", dec_mesh)
    
    # TESTING OUT FILTERING 
    # https://www.open3d.org/docs/latest/tutorial/Basic/mesh.html#Average-filter
    # mesh_out = bpa_mesh.filter_smooth_simple(number_of_iterations=1)
    # mesh_out.compute_vertex_normals()
    # o3d.visualization.draw_geometries([mesh_out])

    # Generate LoDs
    lod_mesh_export(bpa_mesh, [100000, 50000, 10000, 1000, 100], ".ply", output_path)
 

def create_poisson_mesh(pcd, output_path):
    # Poisson surface reconstruction method
    poisson_mesh = o3d.geometry.TriangleMesh.create_from_point_cloud_poisson(pcd, depth=8, width=0, scale=1.1, linear_fit=False)[0]
    poisson_mesh.remove_degenerate_triangles()
    poisson_mesh.remove_duplicated_triangles()
    poisson_mesh.remove_duplicated_vertices()
    poisson_mesh.remove_non_manifold_edges()

    bbox = pcd.get_axis_aligned_bounding_box()
    p_mesh_crop = poisson_mesh.crop(bbox)

    # Save the mesh
    o3d.io.write_triangle_mesh(output_path+"p_mesh_c.ply", p_mesh_crop)

    # Generate LoDs
    lod_mesh_export(p_mesh_crop, [100000, 50000, 10000, 1000, 100], ".ply", output_path)