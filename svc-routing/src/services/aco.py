import time
import numpy as np

def ant_colony_optimization(start_idx, points, dist_matrix):
    start_time = time.time()
    num_points = len(points)
    
    safe_dist_matrix = np.copy(dist_matrix)
    for i in range(num_points):
        for j in range(num_points):
            if i == j or safe_dist_matrix[i][j] == 0:
                safe_dist_matrix[i][j] = 1e-5
            elif safe_dist_matrix[i][j] is None or np.isnan(safe_dist_matrix[i][j]):
                safe_dist_matrix[i][j] = 999999.0

    num_ants = 40        
    num_iterations = 100  
    alpha = 1.0 
    beta = 2.0 
    evaporation_rate = 0.5 
    Q = 100.0

    pheromones = np.ones((num_points, num_points))
    visibility = 1.0 / safe_dist_matrix

    best_path = None
    best_distance = float('inf')

    for iteration in range(num_iterations):
        all_paths = []
        all_distances = []

        # Każda mrówka buduje swoją trasę
        for ant in range(num_ants):
            path = [start_idx]
            unvisited = set(range(num_points))
            unvisited.remove(start_idx)

            current_node = start_idx
            path_distance = 0.0

            # Mrówka podróżuje, aż odwiedzi wszystkie adnotacje
            while unvisited:
                probabilities = []
                unvisited_list = list(unvisited)

                # Obliczanie prawdopodobieństw przejścia
                for next_node in unvisited_list:
                    tau = pheromones[current_node][next_node] ** alpha
                    eta = visibility[current_node][next_node] ** beta
                    probabilities.append(tau * eta)

                probabilities = np.array(probabilities)
                prob_sum = probabilities.sum()

                if prob_sum == 0:
                    probabilities = np.ones(len(probabilities)) / len(probabilities)
                else:
                    probabilities = probabilities / prob_sum

                # Losowy wybór następnego węzła bazując na prawdopodobieństwie
                next_node = np.random.choice(unvisited_list, p=probabilities)

                # Przejście do węzła
                path.append(next_node)
                path_distance += safe_dist_matrix[current_node][next_node]
                unvisited.remove(next_node)
                current_node = next_node

            path_distance += safe_dist_matrix[current_node][start_idx]
            path.append(start_idx)

            all_paths.append(path)
            all_distances.append(path_distance)

            if path_distance < best_distance:
                best_distance = path_distance
                best_path = path

        # Aktualizacja feromonów
        pheromones *= (1.0 - evaporation_rate)

        # Zostawianie nowego śladu przez mrówki
        for path, distance in zip(all_paths, all_distances):
            pheromone_to_add = Q / distance
            for i in range(len(path) - 1):
                u = path[i]
                v = path[i+1]
                pheromones[u][v] += pheromone_to_add
                pheromones[v][u] += pheromone_to_add

    end_time = time.time()
    
    if best_path is None:
        return [], end_time - start_time

    return [int(idx) for idx in best_path], end_time - start_time