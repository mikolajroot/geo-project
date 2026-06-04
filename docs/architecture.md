# Diagram architektury

```mermaid
graph LR
  subgraph proxy-net
    user["Użytkownik"] -->|HTTP| nginx["NGINX (reverse proxy)"]
  end

  subgraph app-net
    nginx -->|proxy-net| svc_auth[svc-auth]
    nginx -->|proxy-net| svc_layers[svc-layers]
    nginx -->|proxy-net| svc_features[svc-features]
    nginx -->|proxy-net| svc_analytics[svc-analytics]
    nginx -->|proxy-net| svc_annotations[svc-annotations]
    nginx -->|proxy-net| svc_revisions[svc-revisions]
  end

  subgraph db-net
    postgres[("Postgres / PostGIS")]
    mongo[("MongoDB")]
    geo_volume["geo_postgres_data (volume)"]
    mongo_volume["mongodb_data (volume)"]
  end

  migration_worker["migration-worker"]

  %% Connections to databases
  svc_auth -->|db-net| postgres
  svc_layers -->|db-net| postgres
  svc_features -->|db-net| postgres
  svc_analytics -->|db-net| postgres
  svc_annotations -->|db-net| postgres
  svc_revisions -->|db-net| postgres
  migration_worker -->|db-net| postgres


  svc_annotations -->|db-net| mongo
  svc_revisions -->|db-net| mongo

  %% Volumes
  postgres --> geo_volume
  mongo --> mongo_volume


  classDef db fill:#ffe6e6,stroke:#ffcccc;
  class postgres,mongo db

  classDef svc fill:#e6f2ff,stroke:#cce0ff;
  class svc_auth,svc_layers,svc_features,svc_analytics,svc_annotations,svc_revisions,migration_worker svc

```