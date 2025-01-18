from flask import Flask, request, jsonify
import os
from flask_cors import CORS
from datetime import datetime
import random
import string

app = Flask(__name__)
CORS(app)

SCREENSHOT_DIR = "screenshots"
AUDIO_DIR = "audio"
os.makedirs(SCREENSHOT_DIR, exist_ok=True)
os.makedirs(AUDIO_DIR, exist_ok=True)

def generate_meeting_id(length=8):
    """Generate a random string of letters and digits as the meeting ID."""
    return ''.join(random.choices(string.ascii_uppercase + string.digits, k=length))

@app.route('/generate-meeting-id', methods=['GET'])
def generate_meeting_id_endpoint():
    meeting_id = generate_meeting_id()
    return jsonify({'meeting_id': meeting_id})

@app.route('/capture-screenshots', methods=['POST'])
def capture_screenshots():
    if 'screenshot' not in request.files:
        return jsonify({'error': 'No screenshot file found in the request'}), 400

    screenshot_file = request.files['screenshot']
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    filename = f"screenshot_{timestamp}.png"
    filepath = os.path.join(SCREENSHOT_DIR, filename)

    try:
        screenshot_file.save(filepath)
        return jsonify({'message': 'Screenshot saved successfully', 'filename': filename}), 200
    except Exception as e:
        return jsonify({'error': str(e)}), 500

@app.route('/capture-audio', methods=['POST'])
def capture_audio():
    if 'audio' not in request.files:
        return jsonify({'error': 'No audio file found in the request'}), 400

    audio_file = request.files['audio']
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    filename = f"audio_{timestamp}.webm"
    filepath = os.path.join(AUDIO_DIR, filename)

    try:
        audio_file.save(filepath)
        return jsonify({'message': 'Audio saved successfully', 'filename': filename}), 200
    except Exception as e:
        return jsonify({'error': str(e)}), 500

if __name__ == '__main__':
    app.run(port=8080, debug=True)
