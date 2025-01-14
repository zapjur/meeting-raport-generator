FROM python:3.9-slim

WORKDIR /app

RUN apt-get update && apt-get install -y \
    gcc g++ python3-dev libtesseract-dev libleptonica-dev tesseract-ocr && \
    rm -rf /var/lib/apt/lists/*

COPY requirements.txt .

RUN pip3 install --upgrade pip
RUN pip3 install --no-cache-dir -r requirements.txt

COPY . .

CMD ["python", "src/main.py"]
