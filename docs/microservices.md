---
layout: default
title: Microservices
nav_order: 2
---

# Microservices Overview

The application consists of several microservices, each with a distinct responsibility.

## Orchestrator
- **Purpose**: Coordinates tasks across all microservices.
- **Key Functionality**:
    - Generates `meetingID` for each meeting.
    - Uses RabbitMQ to distribute tasks to transcription, ocr, summary, report generator, mailer services.
- **API Endpoints**:
    - `GET /generate-meeting-id?email={email}`
        - Initializes a new meeting and returns a `meetingID`.
        - Saves the `meetingID` in Redis with status `started`.
        - Saves user email in Redis with key `meetingID`.

## Transcription Service
- **Purpose**: Converts audio fragments into text and identify speakers.
- **Implementation**:
    - **Uses**: 
        - `pyannote/diarization`, `pyannote/segmentation`, `pyannote/audio`, `pyannote/core` and `pyannote/embeddings` for identifying speakers.
        - `whisper` for transcribing audio.
    
- **Input**: Audio fragment (base64 encoded).
- **Output**: JSON object with transcribed text.

## Summary Service
- **Purpose**: Generates concise summaries of transcriptions.
- **Key Features**:
    - Groups transcription fragments into chunks.
    - Uses LLM (e.g., OpenAI's API) for summarization.
- **Output**: Text summary stored in MongoDB.

## Report Generator
- **Purpose**: Converts transcription, summaries, ocr results, screenshots into PDF report.
- **Key Features**:
    - Formats text into structured sections.
    - Stores generated PDFs in a shared volume.
