import pandas as pd

from aqueduct import op


@op
def sentiment_prediction_using_transformers(reviews):
    import transformers

    model = transformers.pipeline("sentiment-analysis")
    return reviews.join(pd.DataFrame(model(list(reviews["review"]))))
