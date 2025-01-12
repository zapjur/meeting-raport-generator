FROM python:3.9-slim

RUN apt-get update && \
    apt-get install -y ffmpeg && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY requirements.txt .

RUN if [ "$(uname)" = "Darwin" ] || [ "$DOCKER_DEFAULT_PLATFORM" = "linux/arm64" ]; then \
        echo "Installing PyTorch with MPS support..."; \
        pip install torch torchvision torchaudio --index-url https://download.pytorch.org/whl/nightly/cpu; \
    elif [ "$(uname)" = "Linux" ] && nvidia-smi > /dev/null 2>&1; then \
        echo "Installing PyTorch with CUDA support..."; \
        pip install --no-cache-dir torch torchvision torchaudio --extra-index-url https://download.pytorch.org/whl/cu118; \
    else \
        echo "Installing PyTorch for CPU..."; \
        pip install --no-cache-dir torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cpu; \
    fi



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
