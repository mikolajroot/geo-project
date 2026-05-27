import requests
import numpy as np
from typing import List

def get_distance_matrix(points: List) -> np.ndarray:
    coords = ";".join([f"{p.lng},{p.lat}" for p in points])
    url = f"http://router.project-osrm.org/table/v1/driving/{coords}?annotations=distance"
    
    response = requests.get(url)
    if response.status_code == 200:
        data = response.json()
        return np.array(data['distances'])
    else:
        raise Exception(f"Error API OSRM: {response.status_code} - {response.text}")