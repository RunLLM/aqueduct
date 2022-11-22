from constants import SHORT_SENTIMENT_SQL_QUERY
from utils import delete_flow, generate_new_flow_name, run_flow_test, publish_flow_test

from aqueduct import LoadUpdateMode, op


def test_list_saved_objects(client, flow_name, data_integration, engine):
    integration = client.integration(name=data_integration)

    table = integration.sql(query=SHORT_SENTIMENT_SQL_QUERY)
    table.save(integration.config(table="table_1", update_mode=LoadUpdateMode.REPLACE))

    # This will create the table.
    flow = publish_flow_test(
        client,
        name=flow_name(),
        artifacts=table,
        engine=engine,
    )

    # Change to append mode.
    table.save(integration.config(table="table_1", update_mode=LoadUpdateMode.APPEND))
    publish_flow_test(
        client,
        existing_flow=flow,
        artifacts=table,
        engine=engine,
    )

    # Redundant append mode change.
    table.save(integration.config(table="table_1", update_mode=LoadUpdateMode.APPEND))
    publish_flow_test(
        client,
        existing_flow=flow,
        artifacts=table,
        engine=engine,
    )

    # Create a different table from the same artifact.
    table.save(integration.config(table="table_2", update_mode=LoadUpdateMode.REPLACE))
    publish_flow_test(
        client,
        existing_flow=flow,
        artifacts=table,
        engine=engine,
    )

    data = client.flow(flow.id()).list_saved_objects()

    # Check all in same integration
    assert len(data.keys()) == 1

    # table_name, update_mode in order of latest created
    data_set = [
        ("table_2", LoadUpdateMode.REPLACE),
        ("table_1", LoadUpdateMode.APPEND),
        ("table_1", LoadUpdateMode.REPLACE),
    ]
    integration_name = list(data.keys())[0]
    assert len(data[integration_name]) == 3
    for i in range(3):
        item = data[integration_name][i]
        assert (item.object_name, item.update_mode) == data_set[i]

    # Check mapping can be accessed correctly
    # Can be accessed by string of integration name
    assert len(data[data_integration]) == 3

    # Can be accessed by Integration object with integration name
    assert len(data[integration]) == 3


def test_multiple_artifacts_saved_to_same_integration(client, flow_name, data_integration, engine):
    integration = client.integration(name=data_integration)

    table_1 = integration.sql(query=SHORT_SENTIMENT_SQL_QUERY)
    table_1.save(integration.config(table="table_1", update_mode=LoadUpdateMode.REPLACE))
    table_2 = integration.sql(query=SHORT_SENTIMENT_SQL_QUERY)
    table_2.save(integration.config(table="table_2", update_mode=LoadUpdateMode.REPLACE))

    flow = publish_flow_test(
        client,
        name=flow_name(),
        artifacts=[table_1, table_2],
        engine=engine,
    )

    data = client.flow(flow.id()).list_saved_objects()

    assert len(data.keys()) == 1
    data_set = {
        ("table_1", LoadUpdateMode.REPLACE),
        ("table_2", LoadUpdateMode.REPLACE),
    }

    integration_name = list(data.keys())[0]
    assert len(data[integration_name]) == 2
    assert (
        set([(item.object_name, item.update_mode) for item in data[integration_name]])
        == data_set
    )


def test_lazy_artifact_with_save(client, flow_name, data_integration, engine):
    db = client.integration(data_integration)
    reviews = db.sql(SHORT_SENTIMENT_SQL_QUERY)

    @op()
    def copy_field(df):
        df["new"] = df["review"]
        return df

    review_copied = copy_field.lazy(reviews)
    review_copied.save(
        config=db.config(table="test_timestamp_succeeded", update_mode=LoadUpdateMode.REPLACE)
    )

    publish_flow_test(
        client,
        review_copied,
        name=flow_name(),
        engine=engine,
    )
