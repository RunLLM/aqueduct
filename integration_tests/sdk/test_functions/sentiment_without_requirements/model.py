from aqueduct import op

import nltk
from nltk.sentiment.vader import SentimentIntensityAnalyzer


@op()
def sentiment_model_without_requirements(df):
    nltk.download("vader_lexicon")
    sid = SentimentIntensityAnalyzer()
    model = lambda sentence: sid.polarity_scores(sentence)["compound"]
    predictions = df["review"].apply(model)
    df["positivity"] = predictions
    return df
