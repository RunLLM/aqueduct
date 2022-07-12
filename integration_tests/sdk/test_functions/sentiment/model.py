import nltk
from nltk.sentiment.vader import SentimentIntensityAnalyzer

from aqueduct import op


@op
def sentiment_model(df):
    nltk.download("vader_lexicon")
    sid = SentimentIntensityAnalyzer()
    model = lambda sentence: sid.polarity_scores(sentence)["compound"]
    predictions = df["review"].apply(model)
    df["positivity"] = predictions
    return df


@op()
def sentiment_model_multiple_input(df1, df2):
    nltk.download("vader_lexicon")
    sid = SentimentIntensityAnalyzer()
    model = lambda sentence: sid.polarity_scores(sentence)["compound"]
    predictions = df1["review"].apply(model)
    predictions_2 = df2["review"].apply(model)
    df1["positivity"] = predictions
    df1["positivity_2"] = predictions_2
    return df1
