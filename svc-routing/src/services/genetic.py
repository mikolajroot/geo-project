import time
import math
import numpy as np
from sko.GA import GA_TSP

def genetic_algorithm(start_idx, points, dist_matrix):
    start_time = time.time()
    num_points = len(points)
    

    def calc_total_distance(routine):
        dist = 0.0
        for i in range(num_points):
            from_idx = routine[i]
            to_idx = routine[(i + 1) % num_points] 
            
            weight = dist_matrix[from_idx][to_idx]
            
            if weight is None or math.isnan(weight):
                dist += 999999.0 
            else:
                dist += weight
        return dist

    ga_tsp = GA_TSP(
        func=calc_total_distance, 
        n_dim=num_points, 
        size_pop=50, 
        max_iter=300, 
        prob_mut=0.1
    )
    
    best_points, best_distance = ga_tsp.run()
    
   
    best_points = list(best_points)
    start_pos = best_points.index(start_idx)
    
    rotated_path = best_points[start_pos:] + best_points[:start_pos]
    
    rotated_path.append(start_idx)
    
    end_time = time.time()
    
    return [int(idx) for idx in rotated_path], end_time - start_time