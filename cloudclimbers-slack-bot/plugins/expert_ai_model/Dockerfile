FROM python:3.9-slim

WORKDIR /app

COPY expert_ai_model.py .

RUN pip install Flask requests

EXPOSE 8084

CMD ["python", "expert_ai_model.py"]
