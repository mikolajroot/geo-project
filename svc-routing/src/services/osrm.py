import polyline
import requests
import numpy as np
from typing import List

import urllib.parse

def get_distance_matrix(points: List) -> np.ndarray:
    coords = [(p.lat, p.lng) for p in points]
    
    encoded_polyline = polyline.encode(coords, 5)
    
    safe_polyline = urllib.parse.quote(encoded_polyline)
    
    url = f"http://router.project-osrm.org/table/v1/driving/polyline({safe_polyline})?annotations=distance"
    
    response = requests.get(url)
    if response.status_code == 200:
        data = response.json()
        return np.array(data['distances'])
    else:
        raise Exception(f"Error API OSRM: {response.status_code} - {response.text}")