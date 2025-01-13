FROM python:3.9-slim

RUN apt-get update && apt-get install -y ffmpeg libsndfile1 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY requirements.txt .

RUN pip3 install --upgrade pip

# Instalacja PyTorch i zależności
RUN pip3 install torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cpu

# Instalacja starszej wersji importlib_metadata i pyannote.database
RUN pip3 install importlib_metadata==4.13.0 pyannote.database==4.1.1

# Instalacja pozostałych zależności
RUN pip3 install --no-cache-dir -r requirements.txt

# Preload modeli Whisper oraz Pyannote
# RUN python3 -c "import whisper; whisper.load_model('medium')"
#RUN python3 -c "from pyannote.audio import Pipeline; \
#    Pipeline.from_pretrained('pyannote/speaker-diarization'); \
#    Pipeline.from_pretrained('pyannote/segmentation'); \
#    Pipeline.from_pretrained('pyannote/embedding')"

COPY . .

CMD ["python", "src/main.py"]
