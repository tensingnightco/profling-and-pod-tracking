FROM python:3.11

WORKDIR /app

# Install Pyroscope CLI
RUN apt-get update && apt-get install -y curl && \
    curl -L https://github.com/grafana/pyroscope/releases/latest/download/pyroscope_linux_amd64 -o /usr/local/bin/pyroscope && \
    chmod +x /usr/local/bin/pyroscope && \
    rm -rf /var/lib/apt/lists/*

# Install debug symbols
RUN apt-get update && apt-get install -y python3-dbg && rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY app.py .

RUN addgroup --system app && adduser --system --group app
USER app

ENV PORT=8080
ENV PYTHONDEBUG=1
ENV PYTHONFAULTHANDLER=1

EXPOSE 8080

# Use pyroscope exec to automatically instrument
CMD ["pyroscope", "exec", "python", "-u", "app.py"]