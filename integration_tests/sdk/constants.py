# Parameters for the sentiment dataset.
SENTIMENT_SQL_QUERY = "select * from hotel_reviews"
# This is to speed up the database writes.
SHORT_SENTIMENT_SQL_QUERY = "select * from hotel_reviews limit 1"

# Parameters for the churn dataset.
CHURN_SQL_QUERY = "select * from customer_activity"

WINE_SQL_QUERY = "select * from wine"
