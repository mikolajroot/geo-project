from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from src.api.routes import router as optimize_router


app = FastAPI(
    title="Spatial Optimization AI Service",
    description="Mikroserwis AI do optymalizacji tras dla geo-project",
    version="1.0.0"
)


app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


app.include_router(optimize_router, prefix="/api/v1")

@app.get("/health", tags=["System"])
def health_check():
    print("Health check endpoint accessed")
    return {"status": "ok", "service": "routing-ai"}