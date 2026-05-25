from fastapi import APIRouter, HTTPException
import time

from src.api.schema import OptimizationRequest, OptimizationResponse

router = APIRouter(tags=["Optimization"])

@router.post("/optimize", response_model=OptimizationResponse)
def optimize_route(request: OptimizationRequest):
    if len(request.points) < 2:
        raise HTTPException(status_code=400, detail="Wymagane są minimum 2 punkty do optymalizacji.")

    start_time = time.time()

    ordered_ids = [p.id for p in request.points]
    simulated_distance = 0.0 
    # ---------------------------------------------------------------

    end_time = time.time()
    computation_time = (end_time - start_time) * 1000

    return OptimizationResponse(
        ordered_ids=ordered_ids,
        total_distance_km=simulated_distance,
        computation_time_ms=computation_time
    )