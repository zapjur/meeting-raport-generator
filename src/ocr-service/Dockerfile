FROM python:3.10-slim

WORKDIR /app

RUN apt-get update && apt-get install -y \
    gcc g++ python3-dev libtesseract-dev libleptonica-dev tesseract-ocr tesseract-ocr-pol && \
    rm -rf /var/lib/apt/lists/*

ENV TESSDATA_PREFIX=/usr/share/tesseract-ocr/5/tessdata/

COPY requirements.txt .

RUN pip3 install --upgrade pip
RUN pip3 install --no-cache-dir -r requirements.txt

COPY . .

CMD ["python", "src/main.py"]
