[Docs](https://zapjur.github.io/meeting-raport-generator/)

# Usage Instructions

Follow these steps to set up and use the Meeting Report Generator.

## Prerequisites
- **Docker**: Ensure Docker is installed for containerized deployment.

## Setting Up Locally
1. Clone the repository:
   ```bash
   git clone https://github.com/zapjur/meeting-report-generator.git
   cd meeting-report-generator
   ```

2. Create a `.env` file in the root directory with the following environment variables:
   ```bash
   GROQ_API_KEY=YOUR_API_KEY
   HF_TOKEN=YOUR_API_KEY
   OPENAI_API_KEY=YOUR_API_KEY
   EMAIL_PASSWORD="YOUR_API_KEY"
   ```

3. Start services using docker-compose:
   ```bash
   docker-compose up --build -d
   ```

4. Access the web interface at `http://localhost:3000`.

