# Geo-project

Geo-project is a geographic information system (GIS) based on microservice architecture, designed to manage vector layers, spatial objects, and user annotations. The system uses a hybrid database architecture that combines a relational PostgreSQL database (with the PostGIS extension) and a document-based MongoDB database.

## Getting Started

1. Copy the environment variables template file:
   `cp .env.example .env`
2. Fill in the environment variables in the `.env` file.
3. Make sure the `secrets/` directory contains the appropriate files with credentials (`db_user.txt`, `db_password.txt`, `secret_key.txt`).
4. Start the system in the terminal using the defined script:
   `make up`
   (This command runs `docker compose up --build -d` in the background, starting the containers along with the appropriate health check mechanisms).

## Architecture and Microservices

The system consists of independent services that communicate over an internal Docker network, hidden behind a single entry point (Nginx).

- `svc-layers` - stores geographic layers. It communicates with the PostgreSQL database using the Goqu tool.
- `svc-features` - stores spatial objects (features) and their geometries, associated with layers. Built using the GORM ORM.
- `svc-annotations` - stores field notes and annotations associated with spatial objects. Uses MongoDB.
- `svc-revisions` - a service for tracking changes to objects. It implements a simulation of the Mongoose library’s behavior in Go (strong schema validation, embedded subdocuments, pre-hook methods).
- `svc-auth` - handles authorization, user management, and sessions (JWT tokens). Based on the Ent framework.
- `svc-analytics` - a dedicated microservice for querying and analyzing spatial data.
- `nginx` - acts as an API Gateway, routing external traffic to the appropriate services based on URL paths.

## Threats and Security

- Input validation: Each service uses rigorous validators built into the Go language (the `validator/v10` package). Invalid geometries, missing identifiers, or unauthorized requests are rejected with HTTP status codes 400/401/403.
- No Stack Trace Leaks: Services implement a standardized error system defined in the `pkg/errors` package. Every internal error (e.g., a panic or SQL error) is captured and mapped to a uniform JSON object without revealing implementation details to the client.
- Data Integrity: An attempt to delete an object (Feature) while associated annotations exist is blocked with an HTTP 409 Conflict status code.
- Query Security: The use of modern drivers ensures the use of parameterized queries, eliminating vulnerabilities to SQL injection and NoSQL injection.

Translated with DeepL.com (free version)
