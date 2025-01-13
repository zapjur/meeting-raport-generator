FROM python:3.10-slim

RUN apt-get update && \
    apt-get install -y ffmpeg && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt

RUN python -c "from pyannote.audio import Pipeline; \
    Pipeline.from_pretrained('pyannote/speaker-diarization'); \
    Pipeline.from_pretrained('pyannote/segmentation'); \
    Pipeline.from_pretrained('pyannote/embedding')"

RUN python -c "import whisper; whisper.load_model('medium')"

COPY . .

CMD ["python", "src/main.py"]