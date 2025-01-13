from dataclasses import dataclass
from typing import Optional, List, Dict
from typing import List

@dataclass
class Transcription:
    speaker_id: str
    transcription: str
    timestamp_start: str
    timestamp_end: str
    meeting_id: str
    id: Optional[str] = None

    def to_dict(self):
        return {  
            "_id": self.id if self.id else None,
            "speaker_id": self.speaker_id,
            "transcription": self.transcription,
            "timestamp_start": self.timestamp_start,
            "timestamp_end": self.timestamp_end,
            "meeting_id": self.meeting_id
        }

    @staticmethod
    def from_dict(data: dict):
        return Transcription(
            id=str(data.get('_id')),
            speaker_id=data['speaker_id'],
            transcription=data['transcription'],
            timestamp_start=data['timestamp_start'],
            timestamp_end=data['timestamp_end'],
            meeting_id=data['meeting_id']
        )
    
@dataclass
class Embedding:
    meeting_id: str
    embeddings: Dict[str, List[float]]
    id: Optional[str] = None

    def to_dict(self):
        return {
            "_id": self.id if self.id else None,
            "meeting_id": self.meeting_id,
            "embeddings": self.embeddings
        }

    @staticmethod
    def from_dict(data: dict):
        return Embedding(
            id=str(data.get('_id')),
            meeting_id=data['meeting_id'],
            embeddings=data['embeddings']
        )
