from pyannote.audio import Pipeline
from pyannote.audio import Model
from pyannote.core import Segment
from pyannote.audio import Audio
from scipy.spatial.distance import cdist
from transciption import transcript
from dotenv import load_dotenv
from data import Embedding
import pika
import torch
import numpy as np
import os
import json
import mongo_client


# LOAD TOKEN
load_dotenv()
HF_TOKEN = os.getenv("HF_TOKEN")


# LOAD MODELS
diarization_pipeline = Pipeline.from_pretrained("pyannote/speaker-diarization", use_auth_token=HF_TOKEN)
embedding_model = Model.from_pretrained(
    model="pyannote/embedding", use_auth_token=HF_TOKEN, checkpoint="pyannote/embedding"
)
reference_embeddings = {}
# LOAD DB CONNS
transcriptions_collection = mongo_client.connect_to_mongo_collection("transcriptions")
embeddings_collection = mongo_client.connect_to_mongo_collection("embeddings")

# LOAD REFERENCE EMBEDDINGS
def get_reference_embeddings(meeting_id):
    global reference_embeddings
    embeddings = mongo_client.find_document(embeddings_collection, {"meeting_id": meeting_id})
    if embeddings is None:
        reference_embeddings = {}
    else:
        reference_embeddings = embeddings.get("embeddings", {})

def extract_speaker_embeddings(audio_file, diarization_result):
    """Oblicza embeddingi mówców dla fragmentów"""
    embeddings = {}
    audio = Audio()  
    audio_duration = audio.get_duration(audio_file) 
    
    for turn, _, speaker in diarization_result.itertracks(yield_label=True):
        start_time = max(turn.start, 0) 
        end_time = min(turn.end, audio_duration)
        
        if start_time >= end_time:

            continue

        segment = Segment(start_time, end_time)
        waveform, sample_rate = audio.crop(audio_file, segment)
        
        if waveform.shape[0] < 512:  
            continue  

        waveform = torch.tensor(waveform).unsqueeze(0)  
        if waveform.shape[1] == 1:  
            waveform = waveform.repeat(1, 1, 512) 
        embedding = embedding_model.forward(waveform)
        
        if speaker not in embeddings:
            embeddings[speaker] = []
        embeddings[speaker].append(embedding.detach().cpu().numpy()) 


    for speaker in embeddings:
        embeddings[speaker] = np.mean(embeddings[speaker], axis=0)
    
    return embeddings

def match_speakers(current_embeddings, reference_embeddings, threshold=0.3):
    """Dopasowuje mówców do referencji na podstawie embeddingów"""
    speaker_mapping = {}
    
    if len(reference_embeddings) == 0:
        for current_speaker, current_embedding in current_embeddings.items():
            reference_embeddings[current_speaker] = current_embedding
            speaker_mapping[current_speaker] = current_speaker
        return speaker_mapping 

    for current_speaker, current_embedding in current_embeddings.items():
        current_embedding = np.expand_dims(current_embedding, axis=0) 
        
        reference_embeddings_2d = np.vstack(list(reference_embeddings.values()))
        
        distances = cdist(
            current_embedding, 
            reference_embeddings_2d, 
            metric="cosine"
        )
        
        min_distance_idx = np.argmin(distances)
        min_distance = distances[0, min_distance_idx]
        
        if min_distance < threshold:
            reference_speaker = list(reference_embeddings.keys())[min_distance_idx]
            speaker_mapping[current_speaker] = reference_speaker
        else:
            new_speaker_id = f"Speaker {len(reference_embeddings) + 1}"
            reference_embeddings[new_speaker_id] = current_embedding.flatten()
            speaker_mapping[current_speaker] = new_speaker_id

    return speaker_mapping

def export_results_to_file(results, output_file):
    """
    Eksportuje wyniki do pliku JSON.
    """
    with open(output_file, "w", encoding="utf-8") as f:
        json.dump(results, f, ensure_ascii=False, indent=4)
    print(f"Results exported to {output_file}")


def process_audio_chunk(audio_file):
    """Przetwarza fragment audio, dopasowując ID mówców do całej rozmowy"""
    global reference_embeddings

    diarization = diarization_pipeline(audio_file)

    current_embeddings = extract_speaker_embeddings(audio_file, diarization)

    mapping = match_speakers(current_embeddings, reference_embeddings)

    normalized_diarization = []
    for turn, _, speaker in diarization.itertracks(yield_label=True):
        if speaker not in mapping:
            new_speaker_id = f"Speaker_{len(mapping) + 1}"
            mapping[speaker] = new_speaker_id

        normalized_speaker = mapping[speaker]
        normalized_diarization.append((turn.start, turn.end, normalized_speaker))

    return normalized_diarization


def get_audio_file_from_volume(filepath, extensions=[".wav"]):
    #TODO POBRAC NAGRANIE Z WOLUMENU 
    return None

def extract_audio_segment(audio_file, start, end):
    """
    Wycinanie fragmentu audio dla danego zakresu czasowego.
    Jeśli end przekracza długość pliku, zostanie ustawione na koniec pliku.
    """
    from pydub import AudioSegment

    audio = AudioSegment.from_wav(audio_file)
    audio_duration = len(audio) / 1000

    if end > audio_duration:
        print(f"Warning: End time ({end:.2f}s) exceeds file duration ({audio_duration:.2f}s). Adjusting to end of file.")
        end = audio_duration

    os.makedirs('subsegments', exist_ok=True)
    segment = audio[start * 1000:end * 1000] 
    segment_path = os.path.join('subsegments', f"segment_{start:.2f}_{end:.2f}.wav")
    segment.export(segment_path, format="wav")
    return segment_path

def get_latest_ts_end(meeting_id):
    latest_transcription = transcriptions_collection.find(
            {"meeting_id": meeting_id}
        ).sort("timestamp_end", -1).limit(1) 
    return latest_transcription[0]["timestamp_end"] if latest_transcription.count() > 0 else 0

def main(meeting_id, file_path):
    global reference_embeddings
    audio_file = get_audio_file_from_volume(file_path)
    diarization_result = process_audio_chunk(audio_file)
    get_reference_embeddings(meeting_id)
    
    for start, end, speaker in diarization_result:
        
        speaker_audio_file = extract_audio_segment(audio_file, start, end)
        latest_timestamp_end = get_latest_ts_end(meeting_id) 
        transcription = transcript(speaker_audio_file, speaker, latest_timestamp_end)
        mongo_client.insert_document(transcriptions_collection, transcription.to_dict())
    
    updated_embedding = Embedding(
        meeting_id=meeting_id,
        embeddings=reference_embeddings
    )
    mongo_client.upsert_embedding(embeddings_collection, updated_embedding)

    
def callback(ch, method, properties, body):
    """Funkcja wywoływana po otrzymaniu wiadomości z kolejki"""
    try:
        message = json.loads(body)
        file_path = message.get("file_path")
        meeting_id = message.get("meeting_id")

        if not file_path or not meeting_id:
            print("Invalid message format. Skipping...")
            return

        print(f"Received task for file: {file_path}, meeting ID: {meeting_id}")
        main(file_path, meeting_id)

        ch.basic_ack(delivery_tag=method.delivery_tag)  # Potwierdzamy przetworzenie wiadomości
    except Exception as e:
        print(f"Error processing message: {e}")
        ch.basic_nack(delivery_tag=method.delivery_tag)  # Informujemy RabbitMQ o błędzie

# Połączenie z RabbitMQ
def start_consumer():
    RABBITMQ_HOST = "amqp://guest:guest@rabbitmq:5672/"
    RABBITMQ_QUEUE = "transcription_queue"

    connection = pika.BlockingConnection(pika.ConnectionParameters(host=RABBITMQ_HOST))
    channel = connection.channel()

    channel.queue_declare(queue=RABBITMQ_QUEUE, durable=True)  

    channel.basic_consume(
        queue=RABBITMQ_QUEUE,
        on_message_callback=callback
    )

    print(f"Waiting for messages in queue '{RABBITMQ_QUEUE}'. Press Ctrl+C to exit.")
    try:
        channel.start_consuming()
    except KeyboardInterrupt:
        print("Stopping consumer...")
        channel.stop_consuming()
        connection.close()

if __name__ == "__main__":
    start_consumer()