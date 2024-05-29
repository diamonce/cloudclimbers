FROM python:3.9-alpine

WORKDIR /app

COPY get_plugin.py .

RUN pip install --no-cache-dir Flask requests google-auth

EXPOSE 8082

CMD ["python", "get_plugin.py"]
