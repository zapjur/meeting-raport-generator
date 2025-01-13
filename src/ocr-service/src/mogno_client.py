from pymongo import MongoClient
import logging

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S"
)

DB_NAME = "database"

def connect_to_mongo_collection(collection_name: str):
   
    client = MongoClient('mongodb://admin:password@mongodb:27017')
    db = client[DB_NAME]
    collection = db[collection_name]
    logging.info(f"Connect with db={DB_NAME} and collection={collection_name} established successfully")
    return collection

def insert_document(collection, document: dict):
    collection.insert_one(document)

