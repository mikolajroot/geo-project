import time
import math
import numpy as np
import pygad

def genetic_algorithm(start_idx, points, dist_matrix):
    start_time = time.time()
    num_points = len(points)
    
    safe_dist_matrix = np.copy(dist_matrix)
    for i in range(num_points):
        for j in range(num_points):
            if safe_dist_matrix[i][j] is None or np.isnan(safe_dist_matrix[i][j]):
                safe_dist_matrix[i][j] = 999999.0

    def fitness_func(ga_instance, solution, solution_idx):
        dist = 0.0
        for i in range(num_points):
            from_idx = int(solution[i])
            to_idx = int(solution[(i + 1) % num_points])
            dist += safe_dist_matrix[from_idx][to_idx]
            
        if dist == 0:
            return float('inf')
        return 1.0 / dist

    gene_space = list(range(num_points))

    ga_instance = pygad.GA(
        num_generations=1000,
        num_parents_mating=100, 
        fitness_func=fitness_func,
        sol_per_pop=200,
        num_genes=num_points,
        gene_type=int,
        gene_space=gene_space,
        allow_duplicate_genes=False,
        mutation_type="swap",
        mutation_probability=0.2,
        crossover_type="single_point",
        suppress_warnings=True
    )
    
    ga_instance.run()
    
    best_solution, best_solution_fitness, best_match_idx = ga_instance.best_solution()
    

    best_points = [int(idx) for idx in best_solution]
    start_pos = best_points.index(start_idx)
    
    rotated_path = best_points[start_pos:] + best_points[:start_pos]
    rotated_path.append(start_idx)

    end_time = time.time()
    
    return rotated_path, end_time - start_time