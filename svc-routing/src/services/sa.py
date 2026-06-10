import time
import math
import random
import numpy as np

from src.services.nearest_neighbor import nearest_neighbor

def simulated_annealing(start_idx, points, dist_matrix):
    start_time = time.time()
    num_points = len(points)
    
    safe_dist = np.copy(dist_matrix)
    for i in range(num_points):
        for j in range(num_points):
            if safe_dist[i][j] is None or np.isnan(safe_dist[i][j]):
                safe_dist[i][j] = 999999.0

    def calc_dist(path):
        d = 0.0
        for i in range(len(path) - 1):
            d += safe_dist[path[i]][path[i+1]]
        return d

    # Inicjalizacja trasy
    nn_path, _ = nearest_neighbor(start_idx, points, dist_matrix)
    current_path = nn_path[:-1]
    current_dist = calc_dist(current_path)

    best_path = list(current_path)
    best_dist = current_dist

    # Hiperparametry chłodzenia
    T = 10000.0
    T_min = 1.0
    cooling_rate = 0.98
    iter_per_temp = 50

    # Proces wyżarzania
    while T > T_min:
        for _ in range(iter_per_temp):
            # Mutacja Swap: Wybieramy 2 losowe węzły
            i, j = random.sample(range(1, num_points), 2)
            
            new_path = list(current_path)
            new_path[i], new_path[j] = new_path[j], new_path[i] 
            
            new_dist = calc_dist(new_path)
            delta = new_dist - current_dist
            
            # Decyzja: Akceptujemy czy odrzucamy?
            if delta < 0:
                # Trasa jest krótsza
                current_path = new_path
                current_dist = new_dist
                
                # Zapis do najlepszego rowiązania
                if current_dist < best_dist:
                    best_path = list(current_path)
                    best_dist = current_dist
            else:
                # Trasa jest dłuższa
                # Prawdopodobieństwo spada wraz z ochładzaniem systemu
                probability = math.exp(-delta / T)
                if random.random() < probability:
                    current_path = new_path
                    current_dist = new_dist
        
        T = T * cooling_rate
        
    end_time = time.time()
    
    return [int(idx) for idx in best_path], end_time - start_time