FROM python:3.14-slim

RUN apt-get update && apt-get install -y wget gcc make \
    && wget http://webhotel4.ruc.dk/~keld/research/LKH-3/LKH-3.0.14.tgz \
    && tar xvfz LKH-3.0.14.tgz \
    && cd LKH-3.0.14 \
    && make \
    && cp LKH /usr/local/bin/ \
    && cd .. \
    && rm -rf LKH-3.0.14 LKH-3.0.14.tgz

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /app

COPY requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt

COPY src/ ./src/

EXPOSE 8086


CMD ["uvicorn", "src.main:app", "--host", "0.0.0.0", "--port", "8086", "--reload", "--reload-dir", "/app/src"]