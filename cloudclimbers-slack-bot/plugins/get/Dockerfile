FROM python:3.9-alpine

WORKDIR /app

COPY get_plugin.py .

RUN pip install Flask
RUN pip install requests

EXPOSE 8082

CMD ["python", "get_plugin.py"]