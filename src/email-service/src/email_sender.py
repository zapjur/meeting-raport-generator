import smtplib
from email.mime.base import MIMEBase
from email.mime.multipart import MIMEMultipart
from email import encoders
from dotenv import load_dotenv
import os
import logging

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S"
)

def send_email(recipient_email, file_path):
    sender_email = "MeetingRaportGenerator@gmail.com"
    load_dotenv()
    password= os.getenv("EMAIL_PASSWORD")
    subject = "Notatki ze spotkania [FULLY AI GENERATED XD]"
    if not os.path.exists(file_path):
            logging.error(f"File not found: {file_path}")
            return "Error: File not found"
    try:
        # Konfiguracja serwera SMTP
        smtp_server = "smtp.gmail.com"
        smtp_port = 587

        # Utwórz obiekt wiadomości
        message = MIMEMultipart()
        message["From"] = sender_email
        message["To"] = recipient_email
        message["Subject"] = subject

        # Dodaj treść e-maila
        with open(file_path, "rb") as attachment_file:
            part = MIMEBase("application", "octet-stream")
            part.set_payload(attachment_file.read())

        # Kodowanie załącznika w Base64
        encoders.encode_base64(part)

        # Ustawienie nagłówków załącznika
        part.add_header(
            "Content-Disposition",
            f"attachment; filename={file_path.split('/')[-1]}",
        )

        # Dodanie załącznika do wiadomości
        message.attach(part)

        # Połącz się z serwerem SMTP
        with smtplib.SMTP(smtp_server, smtp_port) as server:
            server.starttls()  # Szyfrowanie połączenia
            server.login(sender_email, password)  # Logowanie
            server.sendmail(sender_email, recipient_email, message.as_string())  # Wysyłanie wiadomości
            logging.info(f"Email sent to {recipient_email} sent successfully")
        return "Success!"   
    except Exception as e:
        print(f"Wystąpił błąd podczas wysyłania e-maila: {e}")