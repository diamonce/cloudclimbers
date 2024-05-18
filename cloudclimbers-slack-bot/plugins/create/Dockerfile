FROM python:3.9-slim

WORKDIR /app

COPY plugins/create/create_plugin.py .

RUN pip install Flask

EXPOSE 8081

CMD ["python", "create_plugin.py"]
