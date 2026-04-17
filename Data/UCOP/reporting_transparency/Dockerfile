FROM python:3.12-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY scraper.py .

ENV PYTHONUNBUFFERED=1
ENTRYPOINT ["python", "/app/scraper.py"]
CMD ["--help"]
