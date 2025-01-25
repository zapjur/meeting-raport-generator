---
layout: default
title: Microservices
nav_order: 2
---

# Microservices Overview

The application consists of several microservices, each with a distinct responsibility. Below is an overview of all microservices and their roles within the system.

---

## Orchestrator
- **Purpose**: Coordinates tasks across all microservices.
- **Key Functionality**:
    - Generates `meetingID` for each meeting.
    - Uses Redis for managing meeting metadata and task statuses.
    - Distributes tasks to transcription, OCR, summary, report generator, and mailer services via RabbitMQ.
    - Monitors task completion and transitions the meeting process through stages.

---

## Transcription Service
- **Purpose**: Processes audio fragments to generate transcriptions and identify speakers.
- **Implementation**:
    - **Speaker Diarization**: Uses `pyannote/diarization` and related models to identify distinct speakers.
    - **Transcription**: Uses `whisper` for high-accuracy speech-to-text processing.
    - **Speaker Matching**: Updates or retrieves speaker embeddings from MongoDB for consistent speaker identification across fragments.
- **Input**: File path to audio fragment (.wav) and `meeting_id` via RabbitMQ (`transcription_queue`).
- **Output**: Transcriptions stored in MongoDB (`transcriptions` collection) and updated embeddings in the `embeddings` collection.

---

## OCR Service
- **Purpose**: Extracts text from images or screenshots.
- **Implementation**:
    - Uses an OCR model (e.g., OpenAI's OCR model) for text extraction.
- **Input**: Screenshot file path and `meeting_id` provided via RabbitMQ (`ocr_queue`).
- **Output**: OCR results stored in MongoDB (`ocr_results` collection).

---

## Summary Service
- **Purpose**: Generates concise summaries of transcriptions.
- **Implementation**:
    - Fetches transcriptions from MongoDB.
    - Uses LLMs (e.g., OpenAI or LLaMA with Groq) to generate structured summaries.
- **Input**: `meeting_id` via RabbitMQ (`summary_queue`).
- **Output**: Summaries stored in MongoDB (`summaries` collection).

---

## Report Generator
- **Purpose**: Combines processed data into a structured PDF report.
- **Implementation**:
    - Fetches transcriptions, summaries, OCR results, and screenshots from MongoDB and shared volumes.
    - Generates a PDF report with structured sections.
- **Input**: `meeting_id` via RabbitMQ (`report_queue`).
- **Output**: PDF report stored in a shared volume (`shared-report`).

---

## Mailer Service
- **Purpose**: Sends the final PDF report to the user.
- **Implementation**:
    - Gets the user’s email from RabbitMQ message.
    - Retrieves the PDF report from the shared volume.
    - Sends the email with the report attached.
- **Input**: File path to report and email provided via RabbitMQ (`email_queue`).
- **Output**: Email sent to the user with the attached report.

---

## Logger Service
- **Purpose**: Collects and stores logs from various microservices.
- **Implementation**:
    - Listens to log messages on the `logs_queue`.
    - Saves logs to MongoDB (`logs` collection) for debugging, monitoring, and auditing purposes.
- **Input**: Log messages (structured in JSON format) sent by other services via RabbitMQ.
- **Output**: Logs stored in MongoDB (`logs` collection).

---

## Additional Details
Each service is designed to be modular, scalable, and independent, allowing for easy updates and debugging. Services communicate asynchronously using RabbitMQ to ensure reliability and fault tolerance.
