from enum import Enum


class DataObject(str, Enum):
    SENTIMENT = "hotel_reviews"
    CHURN = "customer_activity"
    WINE = "wine"
    CUSTOMERS = "customers"
