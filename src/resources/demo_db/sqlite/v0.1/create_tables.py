from os.path import join
from pathlib import Path

import pandas as pd
from sqlalchemy import create_engine, engine


def create_diabetes_table(engine):
    df = pd.read_csv(
        "https://raw.githubusercontent.com/aqueducthq/aqueduct/main/src/resources/demo_db/sqlite/v0.1/diabetes.csv",
        names=[
            "pregnancies",
            "glucose",
            "diastolic_bp",
            "skin_thickness",
            "2_hr_insulin",
            "bmi",
            "pedigree_fn",
            "age",
            "has_diabetes",
        ],
        header=None,
    )

    df.to_sql(
        "diabetes",
        con=engine,
        index=False,
        if_exists="replace",
    )


def create_housing_table(engine):
    df = pd.read_csv(
        "https://raw.githubusercontent.com/aqueducthq/aqueduct/main/src/resources/demo_db/sqlite/v0.1/house_prices.csv",
        index_col="Id",
    )

    df.to_sql(
        "house_prices",
        con=engine,
        index=True,
        if_exists="replace",
    )


if __name__ == "__main__":
    url = "sqlite:///{database}".format(
        database=join(Path.home(), ".aqueduct", "server", "db", "demo.db"),
    )

    try:
        engine = create_engine(url)
        engine.connect()

        create_diabetes_table(engine)
        create_housing_table(engine)
    finally:
        engine.dispose()
