import time
import lkh

def generate_lkh(start_idx, points, dist_matrix):
    start_time = time.time()
    n = len(points)
    
    tsplib_str = (
        "NAME: OSRM_Routing\n"
        "TYPE: TSP\n"
        f"DIMENSION: {n}\n"
        "EDGE_WEIGHT_TYPE: EXPLICIT\n"
        "EDGE_WEIGHT_FORMAT: FULL_MATRIX\n"
        "EDGE_WEIGHT_SECTION\n"
    )
    

    for i in range(n):
        row_str = []
        for j in range(n):
            val = dist_matrix[i][j]
            if val is None or val != val:
                row_str.append("999999999")
            else:
                row_str.append(str(int(val)))
        tsplib_str += " ".join(row_str) + "\n"
        
    tsplib_str += "EOF\n"
    
    problem = lkh.LKHProblem.parse(tsplib_str)
    

    routes = lkh.solve(solver='LKH', problem=problem, max_trials=100, runs=1,precision=1)
    
    best_route = routes[0] 
    
    path_0_indexed = [node - 1 for node in best_route]
    
    start_pos = path_0_indexed.index(start_idx)
    rotated_path = path_0_indexed[start_pos:] + path_0_indexed[:start_pos]
    
    rotated_path.append(start_idx)
    
    end_time = time.time()
    return rotated_path, end_time - start_time