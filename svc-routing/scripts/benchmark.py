import matplotlib
import requests
import random
import statistics
import matplotlib.pyplot as plt
matplotlib.use('Agg')

BASE_URL = "http://localhost:80/api/v1"
LOGIN_URL = f"{BASE_URL}/auth/login"
OPTIMIZE_URL = f"{BASE_URL}/routing/optimize"

def get_jwt_token():
    credentials = {
        "login": "dasdsadas",
        "password": "adsdsafsdfsfdsfds"
    }
    response = requests.post(LOGIN_URL, json=credentials)
    response.raise_for_status()
    return response.json()["access_token"]

ALGORITHMS = ["nearest_neighbor", "genetic", "ant_colony", "simulated_annealing"]

PROBLEM_SIZES = [10, 20, 30, 40] 
RUNS_PER_ALGO = 3 

def generate_random_points(num_points):
    points = []
    for i in range(num_points):
        lat = 54.35 + random.uniform(-0.1, 0.1)
        lng = 18.64 + random.uniform(-0.1, 0.1)
        points.append({
            "id": f"point_{i}",
            "lat": round(lat, 5),
            "lng": round(lng, 5)
        })
    return points

def run_benchmark():

    token = get_jwt_token()
    headers = {"Authorization": f"Bearer {token}"}

    time_results = {algo: [] for algo in ALGORITHMS}
    distance_results = {algo: [] for algo in ALGORITHMS}

    for size in PROBLEM_SIZES:

        points = generate_random_points(size)
        
        for algo in ALGORITHMS:
            algo_times = []
            algo_distances = []
            
            
            for _ in range(RUNS_PER_ALGO):
                payload = {
                    "feature_id": 1,
                    "algorithm": algo,
                    "start_idx": 1,
                    "goal_idx": None,
                    "points": points
                }
                
                try:
                    response = requests.post(OPTIMIZE_URL, json=payload, timeout=60,headers=headers)
                    response.raise_for_status()
                    data = response.json()
                    
                    algo_times.append(data.get("computation_time_ms", 0))
                    algo_distances.append(data.get("total_distance_km", 0))
                except requests.exceptions.RequestException as e:
                    print(f" [BŁĄD API: {e}]", end="")
            
            if algo_times and algo_distances:
                avg_time = statistics.mean(algo_times)
                avg_dist = statistics.mean(algo_distances)
                time_results[algo].append(avg_time)
                distance_results[algo].append(avg_dist)
                print(f" Śr. Czas: {avg_time:.2f}ms | Śr. Dystans: {avg_dist:.2f}km")
            else:
                time_results[algo].append(0)
                distance_results[algo].append(0)
                print(" [Zakończono niepowodzeniem]")

    return time_results, distance_results

def plot_results(time_results, distance_results):
 
    plt.figure(figsize=(10, 6))
    for algo in ALGORITHMS:
        plt.plot(PROBLEM_SIZES, time_results[algo], marker='o', label=algo.upper(), linewidth=2)
    
    plt.title("Złożoność Czasowa Algorytmów Optymalizacyjnych")
    plt.xlabel("Liczba punktów w trasie (N)")
    plt.ylabel("Średni czas obliczeń [ms]")
    plt.yscale("log")
    plt.legend()
    plt.grid(True, which="both", ls="--")
    plt.savefig("benchmark_czas.png")

    plt.figure(figsize=(10, 6))
    for algo in ALGORITHMS:
        plt.plot(PROBLEM_SIZES, distance_results[algo], marker='s', label=algo.upper(), linewidth=2)
    
    plt.title("Jakość Wyznaczonych Tras (Całkowity Dystans)")
    plt.xlabel("Liczba punktów w trasie (N)")
    plt.ylabel("Średni wyznaczony dystans [km]")
    plt.legend()
    plt.grid(True)
    plt.savefig("benchmark_dystans.png")

if __name__ == "__main__":
    t_res, d_res = run_benchmark()
    plot_results(t_res, d_res)