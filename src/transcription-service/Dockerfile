FROM python:3.9-slim

RUN apt-get update && \
    apt-get install -y ffmpeg && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY requirements.txt .

RUN pip install --no-cache-dir torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cpu

# Instalacja pozostałych zależności
RUN pip install --no-cache-dir -r requirements.txt

# Preload modeli Whisper oraz Pyannote
RUN python -c "import whisper; whisper.load_model('medium')"
RUN python -c "from pyannote.audio import Pipeline; \
    Pipeline.from_pretrained('pyannote/speaker-diarization'); \
    Pipeline.from_pretrained('pyannote/segmentation'); \
    Pipeline.from_pretrained('pyannote/embedding')"

COPY . .

CMD ["python", "src/main.py"]
