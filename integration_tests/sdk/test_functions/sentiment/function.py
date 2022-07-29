import nltk
from nltk.sentiment.vader import SentimentIntensityAnalyzer

from aqueduct import op


def sentiment_model_function(df):
    nltk.download("vader_lexicon")
    sid = SentimentIntensityAnalyzer()
    model = lambda sentence: sid.polarity_scores(sentence)["compound"]
    predictions = df["review"].apply(model)
    df["positivity"] = predictions
    return df