import json
import os
import math


import requests
from fastapi import APIRouter, HTTPException, Request

from src.api.schema import OptimizationRequest, OptimizationResponse, Point
from src.services.osrm import get_distance_matrix
from src.services.astar import a_star, heuristic
from src.services.nearest_neighbor import nearest_neighbor
from src.services.genetic import genetic_algorithm

router = APIRouter(tags=["Optimization"])


def _annotations_service_url() -> str:
    return os.getenv("ANNOTATIONS_SERVICE_URL", "http://svc-annotations:8084")


def _fetch_annotations(feature_id: int, auth_header: str | None) -> list[dict]:
    url = f"{_annotations_service_url().rstrip('/')}/api/v1/annotations"

    headers = {"Accept": "application/json"}
    if auth_header:
        headers["Authorization"] = auth_header

    try:
        response = requests.get(url, params={"features": str(feature_id)}, headers=headers, timeout=10)
        response.raise_for_status()
        payload = response.json()
    except requests.RequestException as err:
        raise HTTPException(status_code=502, detail="failed to reach annotations service") from err
    except json.JSONDecodeError as err:
        raise HTTPException(status_code=502, detail="invalid annotations service response") from err

    data = payload.get("data")
    if not isinstance(data, list):
        raise HTTPException(status_code=502, detail="invalid annotations service response")

    return data


def _extract_points(raw_annotations: list[dict]) -> list[Point]:
    points: list[Point] = []

    for annotation in raw_annotations:
        location = annotation.get("location") or {}
        coordinates = location.get("coordinates") or location.get("Coordinates") or []
        if len(coordinates) != 2:
            continue

        points.append(
            Point(
                id=str(annotation.get("id", "")),
                lng=float(coordinates[0]),
                lat=float(coordinates[1]),
            )
        )

    return points


def _safe_edge_distance(weight, from_point: Point, to_point: Point) -> float:
    try:
        numeric_weight = float(weight)
    except (TypeError, ValueError):
        return heuristic(from_point, to_point)

    if not math.isfinite(numeric_weight):
        return heuristic(from_point, to_point)

    return numeric_weight


@router.post("/optimize", response_model=OptimizationResponse)
def optimize_route(request: OptimizationRequest, incoming_request: Request):
    raw_annotations = _fetch_annotations(request.feature_id, incoming_request.headers.get("authorization"))
    points = _extract_points(raw_annotations)

    if not points:
        raise HTTPException(status_code=404, detail=f"no annotations found for feature_id {request.feature_id}")

    if request.algorithm not in ["A*", "nearest_neighbor", "genetic"]:
        raise HTTPException(status_code=400, detail="unsupported algorithm")
    
    if request.algorithm == "A*" and (request.goal_idx is None):
        raise HTTPException(status_code=400, detail="goal_idx is required for A* algorithm")

    if request.start_idx < 0 or request.start_idx >= len(points):
        raise HTTPException(status_code=400, detail="start_idx out of range")
    if request.goal_idx is not None and (request.goal_idx < 0 or request.goal_idx >= len(points)):
        raise HTTPException(status_code=400, detail="goal_idx out of range")

    start_idx = request.start_idx
    goal_idx = request.goal_idx

    total_distance: float = 0.0

    match request.algorithm:
        case "A*":
            try:
                dist_matrix = get_distance_matrix(points)
            except Exception as err:
                raise HTTPException(status_code=502, detail=str(err))

            ordered_indices, computation_time = a_star(start_idx, goal_idx, points, dist_matrix)
            if not ordered_indices:
                raise HTTPException(status_code=500, detail="A* failed to find a path")

            ordered_points = [points[i] for i in ordered_indices]

            for i in range(len(ordered_indices) - 1):
                from_idx = ordered_indices[i]
                to_idx = ordered_indices[i + 1]
                total_distance += _safe_edge_distance(dist_matrix[from_idx][to_idx], points[from_idx], points[to_idx])

        case "nearest_neighbor":
            try:
                dist_matrix = get_distance_matrix(points)
            except Exception as err:
                raise HTTPException(status_code=502, detail=str(err))

            ordered_indices, computation_time = nearest_neighbor(start_idx, points, dist_matrix)
            
            if not ordered_indices:
                raise HTTPException(status_code=500, detail="Nearest Neighbor failed")

            ordered_points = [points[i] for i in ordered_indices]


            for i in range(len(ordered_indices) - 1):
                from_idx = ordered_indices[i]
                to_idx = ordered_indices[i + 1]
                total_distance += _safe_edge_distance(dist_matrix[from_idx][to_idx], points[from_idx], points[to_idx])

        case "genetic":
            try:
                dist_matrix = get_distance_matrix(points)
            except Exception as err:
                raise HTTPException(status_code=502, detail=str(err))

            ordered_indices, computation_time = genetic_algorithm(start_idx, points, dist_matrix)
            
            if not ordered_indices:
                raise HTTPException(status_code=500, detail="Genetic Algorithm failed")

            ordered_points = [points[i] for i in ordered_indices]

            for i in range(len(ordered_indices) - 1):
                from_idx = ordered_indices[i]
                to_idx = ordered_indices[i + 1]
                total_distance += _safe_edge_distance(dist_matrix[from_idx][to_idx], points[from_idx], points[to_idx])

    return OptimizationResponse(
        ordered_points=ordered_points,
        total_distance_km=total_distance / 1000.0,
        computation_time_ms=round(computation_time * 1000, 2),
    )