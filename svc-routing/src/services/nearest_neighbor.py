import time
import math

def nearest_neighbor(start_idx, points, dist_matrix):
    start_time = time.time()
    n = len(points)
    

    unvisited = set(range(n))
    unvisited.remove(start_idx)
    
    path = [start_idx]
    current = start_idx
    
    while unvisited:
        nearest = None
        min_dist = float('inf')
        
        for neighbor in unvisited:
            weight = dist_matrix[current][neighbor]
            
            if weight is None or math.isnan(weight):
                continue
                
            if weight < min_dist:
                min_dist = weight
                nearest = neighbor
        
        if nearest is None:
            nearest = unvisited.pop()
        else:
            unvisited.remove(nearest)
            
        path.append(nearest)
        current = nearest
        
    path.append(start_idx)
        
    end_time = time.time()
    return path, end_time - start_time