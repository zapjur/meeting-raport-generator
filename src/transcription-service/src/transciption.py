import whisper
from datetime import timedelta
from data import Transcription 
import mongo_client
import torch
import logging

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S"
)

if torch.cuda.is_available():
    device = torch.device("cuda")  # GPU z CUDA (NVIDIA)
elif torch.backends.mps.is_available() and torch.backends.mps.is_built():
    device = torch.device("mps")  # Multi-Process Service (MPS)
else:
    device = torch.device("cpu")

logging.info(f"Using device: {device}")

model = whisper.load_model("medium").to(device)

def transcript(audio_file, speaker_id, latest_timestamp_end, meeting_id):
    """
    Transkrybuje fragment audio i zwraca wynik w postaci instancji Transcription.
    """
    conn = mongo_client.connect_to_mongo_collection("transcriptions")

    result = model.transcribe(audio_file, language="en")
    transcription_text = result["text"]

    # Oblicz długość fragmentu na podstawie audio_file
    from pydub.utils import mediainfo
    audio_info = mediainfo(audio_file)
    
    # Sprawdzamy, czy mediainfo zawiera pole "duration"
    if "duration" in audio_info:
        duration = float(audio_info["duration"])
    else:
        raise ValueError("Nie udało się uzyskać informacji o czasie trwania pliku audio")

    timestamp_start = float(latest_timestamp_end)
    timestamp_end = timestamp_start + duration

    # Tworzymy instancję Transcription
    return Transcription(
        speaker_id=speaker_id,
        transcription=transcription_text,
        timestamp_start=str(timedelta(seconds=int(timestamp_start))),
        timestamp_end=str(timedelta(seconds=int(timestamp_end))),
        meeting_id=meeting_id
    )

