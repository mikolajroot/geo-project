import heapq
import math
import time

def reconstruct_path(came_from, current):
    total_path = [current]
    while current in came_from:
        current = came_from[current]
        total_path.insert(0, current)
    return total_path

def heuristic(p1, p2):
    return math.hypot(p1.lat - p2.lat, p1.lng - p2.lng) * 111000


def _edge_weight(weight, current_point, neighbor_point):
    try:
        numeric_weight = float(weight)
    except (TypeError, ValueError):
        return heuristic(current_point, neighbor_point)

    if not math.isfinite(numeric_weight):
        return heuristic(current_point, neighbor_point)

    return numeric_weight

def a_star(start_idx, goal_idx, points, dist_matrix):
    start_time = time.time()

    open_set = []
    heapq.heappush(open_set, (0, start_idx))
    # open_set_hash = {start_idx}

    came_from = {}

    g_score = {i: float('inf') for i in range(len(points))}
    g_score[start_idx] = 0

    f_score = {i: float('inf') for i in range(len(points))}
    f_score[start_idx] = heuristic(points[start_idx], points[goal_idx])

    while open_set:
        current_f, current = heapq.heappop(open_set)
        # open_set_hash.remove(current)

        if current_f > f_score[current]:
            continue

        if current == goal_idx:
            end_time = time.time()
            return reconstruct_path(came_from, current), end_time - start_time

        for neighbor in range(len(points)):
            if current == neighbor:
                continue

            weight = _edge_weight(dist_matrix[current][neighbor], points[current], points[neighbor])

            tentative_g_score = g_score[current] + weight

            if tentative_g_score < g_score[neighbor]:
        
                came_from[neighbor] = current
                g_score[neighbor] = tentative_g_score
                f_score[neighbor] = tentative_g_score + heuristic(points[neighbor], points[goal_idx])

                heapq.heappush(open_set, (f_score[neighbor], neighbor))


    return [], time.time() - start_time