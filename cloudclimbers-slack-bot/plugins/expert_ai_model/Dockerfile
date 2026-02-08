FROM python:3.12-alpine

WORKDIR /app

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Copy requirements first for better caching
COPY requirements.txt ./

# Install dependencies with pinned versions
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir -r requirements.txt

COPY expert_ai_model.py .

# Run as non-root user
USER appuser

EXPOSE 8084

CMD ["python", "expert_ai_model.py"]
