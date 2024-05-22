FROM python:3.9-slim

WORKDIR /app

COPY get_plugin.py .

RUN pip install Flask

EXPOSE 8082

CMD ["python", "get_plugin.py"]
