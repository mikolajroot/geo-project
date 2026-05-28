FROM python:3.14-slim

WORKDIR /app

COPY index.html /app/index.html

EXPOSE 8087

CMD ["python", "-m", "http.server", "8087"]
