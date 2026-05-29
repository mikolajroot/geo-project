import time
import math
import numpy as np

if not hasattr(np, 'int'):
    np.int = int # type: ignore
    
from sko.ACA import ACA_TSP

def ant_colony_optimization(start_idx, points, dist_matrix):
    start_time = time.time()
    num_points = len(points)
    
    safe_dist_matrix = np.copy(dist_matrix)
    for i in range(num_points):
        for j in range(num_points):
            if i == j or safe_dist_matrix[i][j] == 0:
                safe_dist_matrix[i][j] = 1e-5
            elif safe_dist_matrix[i][j] is None or math.isnan(safe_dist_matrix[i][j]):
                safe_dist_matrix[i][j] = 999999.0

    def calc_total_distance(routine):
        dist = 0.0
        for i in range(num_points):
            from_idx = routine[i]
            to_idx = routine[(i + 1) % num_points]
            dist += safe_dist_matrix[from_idx][to_idx]
        return dist

    aca = ACA_TSP(
        func=calc_total_distance, 
        n_dim=num_points, 
        size_pop=40, 
        max_iter=100, 
        distance_matrix=safe_dist_matrix
    )
    
    best_points, best_distance = aca.run()
    
    best_points = list(best_points)
    start_pos = best_points.index(start_idx)
    
    rotated_path = best_points[start_pos:] + best_points[:start_pos]
    rotated_path.append(start_idx)
    
    end_time = time.time()
    
    return [int(idx) for idx in rotated_path], end_time - start_time