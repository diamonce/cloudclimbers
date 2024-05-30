FROM python:3.9-alpine

WORKDIR /app

COPY plugins/delete/delete_plugin.py .

RUN pip install Flask

EXPOSE 8083

CMD ["python", "delete_plugin.py"]
