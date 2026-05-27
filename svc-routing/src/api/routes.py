import json
import math
import os
import time
from urllib.error import HTTPError, URLError
from urllib.parse import urlencode
from urllib.request import Request as UrlRequest
from urllib.request import urlopen

from fastapi import APIRouter, HTTPException, Request

from src.api.schema import OptimizationRequest, OptimizationResponse

router = APIRouter(tags=["Optimization"])


def _annotations_service_url() -> str:
    return os.getenv("ANNOTATIONS_SERVICE_URL", "http://svc-annotations:8084")


def _fetch_annotations(feature_id: int, auth_header: str | None) -> list[dict]:
    query = urlencode({"features": str(feature_id)})
    url = f"{_annotations_service_url().rstrip('/')}/api/v1/annotations?{query}"

    headers = {"Accept": "application/json"}
    if auth_header:
        headers["Authorization"] = auth_header

    request = UrlRequest(url, headers=headers, method="GET")

    try:
        with urlopen(request, timeout=10) as response:
            payload = json.loads(response.read().decode("utf-8"))
    except HTTPError as err:
        detail = err.read().decode("utf-8", errors="ignore") if err.fp else str(err)
        raise HTTPException(status_code=err.code, detail=detail or "failed to fetch annotations") from err
    except URLError as err:
        raise HTTPException(status_code=502, detail="failed to reach annotations service") from err

    data = payload.get("data")
    if not isinstance(data, list):
        raise HTTPException(status_code=502, detail="invalid annotations service response")

    return data


def _extract_points(raw_annotations: list[dict]) -> list[dict]:
    points: list[dict] = []

    for annotation in raw_annotations:
        location = annotation.get("location") or {}
        coordinates = location.get("Coordinates") or []
        if len(coordinates) != 2:
            continue

        points.append({
            "id": str(annotation.get("id", "")),
            "lng": float(coordinates[0]),
            "lat": float(coordinates[1]),
        })

    return points


@router.post("/optimize", response_model=OptimizationResponse)
def optimize_route(request: OptimizationRequest, incoming_request: Request):
    raw_annotations = _fetch_annotations(request.feature_id, incoming_request.headers.get("authorization"))
    points = _extract_points(raw_annotations)

    if not points:
        raise HTTPException(status_code=404, detail=f"no annotations found for feature_id {request.feature_id}")

    if request.algorithm not in {"aco", "ga", "nearest_neighbor", "A*", "local_search", "simulated_annealing"}:
        raise HTTPException(status_code=400, detail="unsupported algorithm")


    return OptimizationResponse(
        ordered_ids=points,
    )