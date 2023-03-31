from aqueduct.decorator import op

import aqueduct as aq

NAME = "succeed_march_madness_dag_layout_test"
DESCRIPTION = """
    * This test is here so that we can check for DAG positioning issues and edge overlaps.
    * Edges shuold not overlap or be minimized.
    * Nodes should be evenly spaced out and not have extra distance for edges connecting an operator to a check/metric.
"""


@op(requirements=[])
def compile_regular_season_stats(df):
    return df


@op(requirements=[])
def create_team_ranks(ranks):
    return ranks


@op(requirements=[])
def create_training_dataset(tourney_results, team_ranks, reg_season_stats):
    return tourney_results


@op(requirements=[])
def train_forest_model(train):
    return train


@op(requirements=[])
def generate_submission(model, team_ranks, reg_season_stats, test):
    return test


def deploy(client, integration_name):
    aq.global_config({"lazy": True})

    warehouse = client.integration(integration_name)
    m_regular_season_detailed_results = warehouse.sql(
        "select * from hotel_reviews;", name="reg_season_results"
    )

    regular_season_stats_table = compile_regular_season_stats(m_regular_season_detailed_results)

    ranks_table = warehouse.sql("select * from hotel_reviews;", name="ranking_compilation")

    team_ranks_table = create_team_ranks(ranks_table)

    tourney_results_table = warehouse.sql("select * from hotel_reviews;", name="tourney_results")

    train_table = create_training_dataset(
        tourney_results_table, team_ranks_table, regular_season_stats_table
    )

    model = train_forest_model(train_table)

    test_table = warehouse.sql("select * from hotel_reviews;", name="sample_submission")

    submission_table = generate_submission(
        model, team_ranks_table, regular_season_stats_table, test_table
    )

    client.publish_flow(name=NAME, description=DESCRIPTION, artifacts=[submission_table])
