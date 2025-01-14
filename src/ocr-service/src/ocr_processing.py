import pytesseract
import logging
import os
from PIL import Image
import mongo_client
from ocr_result import OCRResult

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S"
)

def ocr_image(file_path, meeting_id):
    try:
        if not os.path.exists(file_path):
            logging.error(f"File not found: {file_path}")
            return "Error: File not found"

        logging.info(f"Performing OCR on image with path={file_path} for meeting with meeting_id={meeting_id}")
        custom_config = r'--psm 6 --oem 3 -c preserve_interword_spaces=1'
        img = Image.open(file_path)

        text = pytesseract.image_to_string(img, config=custom_config, lang='pol')

        logging.info(f"Saving OCR text result in MongoDB for meeting with meeting_id={meeting_id}")
        mongo_conn = mongo_client.connect_to_mongo_collection('ocr_results')
        result = OCRResult(text_result=text, meeting_id=meeting_id)
        mongo_client.insert_document(mongo_conn, result.to_dict())

        logging.info(f"OCR result saved successfully for meeting_id={meeting_id}")
        return "Success"

    except Exception as e:
        logging.error(f"Error while performing OCR on image with path={file_path} for meeting with meeting_id={meeting_id}: {e}", exc_info=True)
        return f"Error: {e}"
