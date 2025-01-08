from pymongo import MongoClient
from data import Embedding
DB_NAME = "database" # we have only one db so it can be stored here

def connect_to_mongo_collection(collection_name: str):
   
    client = MongoClient('mongodb://admin:password@mongodb:27017')
    db = client[DB_NAME]
    collection = db[collection_name]
    return collection

def insert_document(collection, document: dict):
    collection.insert_one(document)

def find_document(collection, query: dict):
    return collection.find_one(query)

def upsert_embedding(collection, embedding: Embedding):
    query = {"meeting_id": embedding.meeting_id}
    update = {
        "$set": embedding.to_dict()
    }
    collection.update_one(query, update, upsert=True)