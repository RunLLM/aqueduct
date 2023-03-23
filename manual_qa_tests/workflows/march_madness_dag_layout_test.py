import aqueduct as aq
from aqueduct.decorator import op, check, metric

NAME = "MarchMadnessDagLayoutTestWorkflow"
DESCRIPTION = """
    Mock of the March Madness Workflow to test out the DAG layout.
    TODO: Add more information here on what to expect from the layout.
"""


@op(requirements=[])
def compile_regular_season_stats(df):
    # import pandas as pd

    # reg_season = df
    # w_cols = ['Season', 'WTeamID', 'WLoc', 'WFGM', 'WFGA', 'WFGM3', 'WFGA3', 'WFTM', 'WFTA', 'WOR', 'WDR', 'WAst', 'WTO', 'WStl', 'WBlk', 'WPF']
    # l_cols = ['Season', 'LTeamID', 'LFGM', 'LFGA', 'LFGM3', 'LFGA3', 'LFTM', 'LFTA', 'LOR', 'LDR', 'LAst', 'LTO', 'LStl', 'LBlk', 'LPF']
    # cols = ['Season', 'TeamID', 'FGM', 'FGA', 'FGM3', 'FGA3', 'FTM', 'FTA', 'OR', 'DR', 'Ast', 'TO', 'Stl', 'Blk', 'PF']

    # w_stats = reg_season[w_cols].groupby(['Season', 'WTeamID']).mean().reset_index()
    # w_stats.columns = cols

    # l_stats = reg_season[l_cols].groupby(['Season', 'LTeamID']).mean().reset_index()
    # l_stats.columns = cols

    # reg_season_stats = pd.concat([w_stats, l_stats]).groupby(['Season', 'TeamID']).mean().reset_index()
    # reg_season_stats['id'] = reg_season_stats.Season.astype(str) + reg_season_stats.TeamID.astype(str)

    # return reg_season_stats
    return df


@op(requirements=[])
def create_team_ranks(ranks):
    # import pandas as pd

    # ranks_agg = ranks.groupby(['Season', 'TeamID']).agg({'OrdinalRank': ['mean', 'min', 'max']})
    # ranks_agg.columns = ['_'.join(col) for col in ranks_agg.columns]
    #
    # team_ranks = ranks_agg.reset_index()
    # team_ranks['id'] = team_ranks.Season.astype(str) + team_ranks.TeamID.astype(str)

    # return team_ranks
    return ranks


@op(requirements=[])
def create_training_dataset(tourney_results, team_ranks, reg_season_stats):
    # import pandas as pd

    def is_winning(wteam, lteam):
        return 1
        # if wteam < lteam:
        #     return 1
        # else:
        #     return 0

    # train = tourney_results
    #
    # train['is_win'] = train.apply(lambda x: is_winning(x['WTeamID'], x['LTeamID']), axis=1)
    # train['team_a'] = train.Season.astype(str) + train.WTeamID.astype(str)
    # train['team_b'] = train.Season.astype(str) + train.LTeamID.astype(str)
    # train = train.drop(['WScore', 'LScore'], axis=1)
    # train = pd.merge(train, team_ranks, left_on='team_a', right_on='id').merge(team_ranks, left_on='team_b', right_on='id', suffixes=('_teama', '_teamb'))
    # train = train.drop(['Season_x', 'Season_y', 'id_teama', 'id_teamb', 'TeamID_teama', 'TeamID_teamb'], axis=1)
    # train = pd.merge(train, reg_season_stats, left_on='team_a', right_on='id').merge(reg_season_stats, left_on='team_b', right_on='id', suffixes=('_teama', '_teamb'))
    # train = train.drop(['Season_x', 'Season_y', 'team_a', 'team_b', 'TeamID_teama', 'TeamID_teamb', 'id_teama', 'id_teamb'], axis=1)
    # train = train.drop(['DayNum', 'WTeamID', 'LTeamID', 'NumOT', 'Season', 'WLoc'], axis=1)

    # return train
    return tourney_results


@op(requirements=[])
def train_forest_model(train):
    # from sklearn.model_selection import train_test_split

    # X = train.drop(['is_win'], axis=1)
    # y = train.is_win
    #
    # X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)
    #
    # from sklearn.ensemble import RandomForestClassifier
    # from sklearn.metrics import classification_report
    #
    # model = RandomForestClassifier(random_state=42)
    # model.fit(X_train, y_train)
    #
    # y_pred = model.predict(X_test)
    #
    # print(classification_report(y_test, y_pred))

    # return model
    return train


@op(requirements=[])
def generate_submission(model, team_ranks, reg_season_stats, test):
    # import pandas as p

    def return_value(data, data_type=None):
        # data = data.split('_')
        # if data_type == 'Team A':
        #     return str(data[0]) + str(data[1])
        # else:
        #     return str(data[0]) + str(data[2])
        return data

    # test['team_a'] = test.apply(lambda x: return_value(x.ID, 'Team A'), axis=1)
    # test['team_b'] = test.apply(lambda x: return_value(x.ID), axis=1)
    # test = pd.merge(test, team_ranks, left_on='team_a', right_on='id').merge(team_ranks, left_on='team_b', right_on='id', suffixes=('_teama', '_teamb'))
    # test = pd.merge(test, reg_season_stats, left_on='team_a', right_on='id').merge(reg_season_stats, left_on='team_b', right_on='id', suffixes=('_teama', '_teamb'))
    # test = test.drop(['Season_teama', 'Season_teamb', 'team_a', 'team_b', 'TeamID_teama', 'TeamID_teamb', 'id_teama', 'id_teamb'], axis=1)

    # test.head()

    # X = test.drop(['ID', 'Pred'], axis=1)

    # test['Pred'] = model.predict_proba(X)[:, 1]
    # test = test[['ID', 'Pred']]

    return test


def deploy(client, integration_name):
    aq.global_config({'lazy': True})

    warehouse = client.integration(integration_name)
    m_regular_season_detailed_results = warehouse.sql(
        "select * from hotel_reviews;", name="reg_season_results")

    regular_season_stats_table = compile_regular_season_stats(
        m_regular_season_detailed_results)
    # regular_season_stats_table.get()

    ranks_table = warehouse.sql(
        "select * from hotel_reviews;", name="ranking_compilation")
    # ranks_table.get()

    team_ranks_table = create_team_ranks(ranks_table)
    # team_ranks_table.get()

    tourney_results_table = warehouse.sql(
        "select * from hotel_reviews;", name="tourney_results")

    train_table = create_training_dataset(
        tourney_results_table, team_ranks_table, regular_season_stats_table)
    # train_table.get()

    model = train_forest_model(train_table)

    test_table = warehouse.sql(
        "select * from hotel_reviews;", name="sample_submission")
    # test_table.get()

    submission_table = generate_submission(
        model, team_ranks_table, regular_season_stats_table, test_table)

    client.publish_flow(
        name="MarchMadnessDagLayoutTestWorkflow",
        artifacts=[submission_table]
    )
