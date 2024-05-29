# How about this base image, it has smaller size
# FROM python:3.9-alpine

FROM python:3.9-slim

WORKDIR /app

COPY create_plugin.py .

RUN pip install Flask

EXPOSE 8085

CMD ["python", "create_plugin.py"]
