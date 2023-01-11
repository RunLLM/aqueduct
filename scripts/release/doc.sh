GITBOOK_REPO=$HOME/gitbook

cd ~/aqueduct/sdk

bash ../scripts/generate_docs.sh 
cp -r docs/* $GITBOOK_REPO/api-reference/sdk-reference/package-aqueduct
rm -r docs/

cd ~/aqueduct
python3 scripts/convert_nb_to_md.py --input="examples/churn_prediction/Customer Churn Prediction.ipynb" --output=$GITBOOK_REPO/example-workflows/customer-churn-predictor.md
python3 scripts/convert_nb_to_md.py --input="examples/sentiment-analysis/Sentiment Model.ipynb" --output=$GITBOOK_REPO/example-workflows/sentiment-analysis.md
python3 scripts/convert_nb_to_md.py --input="examples/wine-ratings-prediction/Predict Missing Wine Ratings.ipynb" --output=$GITBOOK_REPO/example-workflows/wine-ratings-predictor.md
python3 scripts/convert_nb_to_md.py --input="examples/house-price-prediction/House Price Prediction.ipynb" --output=$GITBOOK_REPO/example-workflows/house-price-prediction.md
python3 scripts/convert_nb_to_md.py --input="examples/mpg-regressor/Predicting MPG.ipynb" --output=$GITBOOK_REPO/example-workflows/mpg-regressor.md
python3 scripts/convert_nb_to_md.py --input="examples/diabetes-classifier/Classifying Diabetes Risk.ipynb" --output=$GITBOOK_REPO/example-workflows/diabetes-classifier.md
python3 scripts/convert_nb_to_md.py --input="examples/tutorials/Quickstart Tutorial.ipynb" --output=$GITBOOK_REPO/example-workflows/quickstart-tutorial.md
python3 scripts/convert_nb_to_md.py --input="examples/tutorials/Parameters Tutorial.ipynb" --output=$GITBOOK_REPO/example-workflows/parameters-tutorial.md

echo "Please commit changes and publish a PR for gitbook repo"