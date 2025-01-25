# System Architecture

The Meeting Report Generator uses a microservices-based architecture to ensure scalability, flexibility, and maintainability.

## Architecture Overview
- **Frontend**: A web interface that interacts with users, sends audio fragments for transcription and screenshots for OCR.
- **Backend**: Multiple microservices written in Go and Python handle transcription, summary generation, and PDF creation.
- **Database**: MongoDB stores transcriptions, summaries, and OCR results.
- **Redis** Stores temporary data and information about status of each task.
- **Orchestrator**: A central service manages task distribution using RabbitMQ.

### Architecture Diagram
![Architecture Diagram](/docs/assets/diagram.jpg)    

## Data Flow
1. The frontend captures audio in real-time using the Media Capture and sends 5 minutes fragments to the backend.
2. The Orchestrator assigns tasks to appropriate microservices:
    - **Transcription Service**: Processes audio into text.
    - **Summary Service**: Groups transcriptions and generates summaries.
    - **PDF Generator**: Creates a structured report.
3. OCR analyzes images from virtual whiteboards or shared screens.
4. MongoDB stores the processed data for further use.
