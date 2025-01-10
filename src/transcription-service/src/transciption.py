import whisper
from datetime import timedelta
from data import Transcription 
import mongo_client

model = whisper.load_model("medium")

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

