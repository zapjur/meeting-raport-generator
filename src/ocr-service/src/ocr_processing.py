import pytesseract
import logging
from PIL import Image
import mogno_client
from ocr_result import OCRResult

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S"
)

# Funkcja do wykonania OCR na obrazie
def ocr_image(file_path, meeting_id):
    try:
        logging.info(f"Performing OCR on image with path={file_path} for meeting with meeting_id={meeting_id}")
        custom_config = r'--psm 6 --oem 3 -c preserve_interword_spaces=1'
        img = Image.open(file_path)
        
        # Wykonywanie OCR
        text = pytesseract.image_to_string(img, config=custom_config, lang="pol")

        # Zapisanie wyniku OCR w mongodb
        logging.info(f"Saving ocr text result in mongo_db for meeting with meeting_id={meeting_id}")
        mongo_conn =  mogno_client.connect_to_mongo_collection('ocr_results')
        result = OCRResult(text=text, meeting_id=meeting_id)
        mogno_client.insert_document(mongo_conn, result.to_dict)

    except Exception as e:
        logging.info(f"Performing OCR on image with path={file_path} for meeting with meeting_id={meeting_id} resulted in Error")
        return f"Error: {e}"

