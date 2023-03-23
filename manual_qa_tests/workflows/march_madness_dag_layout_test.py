import aqueduct as aq
from aqueduct.decorator import op, check, metric

NAME = "MarchMadnessDagLayoutTestWorkflow"
DESCRIPTION = """
    Mock of the March Madness Workflow to test out the DAG layout.
    TODO: Add more information here on what to expect from the layout.
"""


@op(requirements=[])
def compile_regular_season_stats(df):
    return df


@op(requirements=[])
def create_team_ranks(ranks):
    return ranks


@op(requirements=[])
def create_training_dataset(tourney_results, team_ranks, reg_season_stats):

    def is_winning(wteam, lteam):
        return 1

    return tourney_results


@op(requirements=[])
def train_forest_model(train):
    return train


@op(requirements=[])
def generate_submission(model, team_ranks, reg_season_stats, test):
    return test


def deploy(client, integration_name):
    aq.global_config({'lazy': True})

    warehouse = client.integration(integration_name)
    m_regular_season_detailed_results = warehouse.sql(
        "select * from hotel_reviews;", name="reg_season_results")

    regular_season_stats_table = compile_regular_season_stats(
        m_regular_season_detailed_results)

    ranks_table = warehouse.sql(
        "select * from hotel_reviews;", name="ranking_compilation")

    team_ranks_table = create_team_ranks(ranks_table)

    tourney_results_table = warehouse.sql(
        "select * from hotel_reviews;", name="tourney_results")

    train_table = create_training_dataset(
        tourney_results_table, team_ranks_table, regular_season_stats_table)

    model = train_forest_model(train_table)

    test_table = warehouse.sql(
        "select * from hotel_reviews;", name="sample_submission")

    submission_table = generate_submission(
        model, team_ranks_table, regular_season_stats_table, test_table)

    client.publish_flow(
        name="MarchMadnessDagLayoutTestWorkflow",
        artifacts=[submission_table]
    )
