FROM python:3.9-slim

WORKDIR /app

COPY delete_plugin.py .

RUN pip install Flask

EXPOSE 8083

CMD ["python", "delete_plugin.py"]
