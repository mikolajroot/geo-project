from pydantic import BaseModel, Field
from typing import List

class Point(BaseModel):
    id: str
    lat: float = Field(..., description="Szerokość geograficzna (Latitude)")
    lng: float = Field(..., description="Długość geograficzna (Longitude)")

class OptimizationRequest(BaseModel):
    feature_id: int = Field(..., description="ID featura do optymalizacji")
    algorithm: str = Field(default="aco", description="Dostępne: aco, ga, nearest_neighbor, A*, local_search,simulated_annealing")

class OptimizationResponse(BaseModel):
    ordered_ids: list[dict]
    # total_distance_km: float
    # computation_time_ms: float