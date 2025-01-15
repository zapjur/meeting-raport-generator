import logging
import os
from dotenv import load_dotenv
import mongo_client
import base64
from ocr_result import OCRResult
from openai import OpenAI

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S"
)
prompt = """
    Wykonaj rozpoznanie tekstu (OCR) ze zrzutu ekranu, który przedstawia slajd z prezentacji wyświetlonej w aplikacji Teams lub Zoom. 
    Skup się wyłącznie na tekście widocznym na slajdzie prezentacji i zignoruj wszystkie inne elementy ekranu, takie jak paski narzędzi, ikony, czy interfejs aplikacji. Zwróć szczególną uwagę na:

    Dokładne odwzorowanie treści tekstu, w tym wszelkich znaków specjalnych, formatowania (np. wypunktowania, nagłówków).
    Rozróżnienie między treścią główną a dodatkowymi notatkami, jeśli są widoczne.
    Wykluczenie wszystkich zbędnych danych spoza obszaru slajdu.
    Wykluczenie takich danych jak numer strony i tym podobne informacje na slajdzie nie niosące żadnej wartości merytorycznej. 
    Wynik powinien być zapisany w strukturze tekstowej, która odzwierciedla strukturę logiczną i hierarchiczną prezentacji. Jeśli jest to możliwe, odwzoruj układ w sposób czytelny dla człowieka.
"""
def create_gpt_client():
    # loading gpt api key from .env
    load_dotenv()
    OPENAI_API_KEY = os.getenv("OPENAI_API_KEY")
    if not OPENAI_API_KEY:
        raise ValueError("OPENAI_API_KEY is not set. Please check your environment variables.")

    client = OpenAI(api_key=OPENAI_API_KEY)
    return client

def encode_image(image_path):
    with open(image_path, "rb") as image_file:
        return base64.b64encode(image_file.read()).decode('utf-8')

def ocr_image(file_path, meeting_id):
    try:
        if not os.path.exists(file_path):
            logging.error(f"File not found: {file_path}")
            return "Error: File not found"

        logging.info(f"Performing OCR on image with path={file_path} for meeting with meeting_id={meeting_id}")
        base64_image = encode_image(file_path)
        client = create_gpt_client()
        response = client.chat.completions.create(
        model="gpt-4o",
        messages=[
                    {
                        "role": "user",
                        "content": [
                            {"type": "text", "text": prompt},
                            {
                                "type": "image_url",
                                "image_url": {
                                    "url": f"data:image/jpeg;base64,{base64_image}",
                                },
                            },
                        ],
                    }
                ],
            )
        logging.info(f"Saving OCR text result in MongoDB for meeting with meeting_id={meeting_id}")
        mongo_conn = mongo_client.connect_to_mongo_collection('ocr_results')
        result = OCRResult(text_result=response.choices[0].message.content, meeting_id=meeting_id)
        mongo_client.insert_document(mongo_conn, result.to_dict())

        logging.info(f"OCR result saved successfully for meeting_id={meeting_id}")
        return "Success"

    except Exception as e:
        logging.error(f"Error while performing OCR on image with path={file_path} for meeting with meeting_id={meeting_id}: {e}", exc_info=True)
        return f"Error: {e}"
