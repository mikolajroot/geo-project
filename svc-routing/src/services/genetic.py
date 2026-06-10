import random
import time
import numpy as np
import pygad
from src.services.nearest_neighbor import nearest_neighbor

def pmx_crossover(parents, offspring_size, ga_instance):
    offspring = np.empty(offspring_size)
    
    for i in range(offspring_size[0]):
        parent1 = parents[i % parents.shape[0]].copy()
        parent2 = parents[(i + 1) % parents.shape[0]].copy()
        
        size = len(parent1)
        child = np.full(size, -1.0) 
        
        cxpoint1, cxpoint2 = sorted(np.random.choice(range(size), 2, replace=False))
        

        child[cxpoint1:cxpoint2] = parent1[cxpoint1:cxpoint2]
        

        for idx in range(cxpoint1, cxpoint2):
            val2 = parent2[idx]

            if val2 not in child:
                curr_idx = idx

                while cxpoint1 <= curr_idx < cxpoint2:
                    val1 = parent1[curr_idx]
                    curr_idx = np.where(parent2 == val1)[0][0]
                
                child[curr_idx] = val2
                

        for j in range(size):
            if child[j] == -1.0:
                child[j] = parent2[j]
                
        offspring[i] = child
        
    return offspring

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

    nn_path, _ = nearest_neighbor(start_idx, points, dist_matrix)
    
    super_chromosome = nn_path[:-1] 
    
    sol_per_pop = 100
    initial_pop = [super_chromosome]
    
    for _ in range(sol_per_pop - 1):
        random_chrom = list(range(num_points))
        random.shuffle(random_chrom)
        initial_pop.append(random_chrom)

    ga_instance = pygad.GA(
        num_generations=400,
        num_parents_mating=100, 
        fitness_func=fitness_func,
        initial_population=initial_pop,
        num_genes=num_points,
        gene_type=int,
        gene_space=gene_space,
        allow_duplicate_genes=False,
        mutation_type="swap",
        mutation_probability=0.02,
        crossover_type=pmx_crossover,  # type: ignore
        keep_elitism=2,
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