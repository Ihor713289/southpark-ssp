FROM python:3.13-slim

ENV PYTHONUNBUFFERED=1

WORKDIR /app

COPY producer/requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt

COPY producer/app.py producer/quotes.json ./

CMD ["python", "app.py"]
