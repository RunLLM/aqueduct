import pandas as pd
from aqueduct.constants.enums import LoadUpdateMode
from constants import SHORT_SENTIMENT_SQL_QUERY
from utils import generate_table_name, publish_flow_test, save

from aqueduct import op


def test_list_saved_objects(client, flow_name, data_integration, engine, validator):
    integration = client.integration(name=data_integration)

    table_1_save_name = generate_table_name()
    table_2_save_name = generate_table_name()

    table = integration.sql(query=SHORT_SENTIMENT_SQL_QUERY)
    extracted_table_data = table.get()
    save(integration, table, name=table_1_save_name, update_mode=LoadUpdateMode.REPLACE)

    # This will create the table.
    flow = publish_flow_test(
        client,
        name=flow_name(),
        artifacts=table,
        engine=engine,
    )
    validator.check_saved_artifact(flow, table.id(), expected_data=extracted_table_data)

    # Change to append mode.
    save(integration, table, name=table_1_save_name, update_mode=LoadUpdateMode.APPEND)
    publish_flow_test(
        client,
        existing_flow=flow,
        artifacts=table,
        engine=engine,
    )
    validator.check_saved_artifact(
        flow,
        table.id(),
        expected_data=pd.concat([extracted_table_data, extracted_table_data], ignore_index=True),
    )

    # Redundant append mode change
    save(integration, table, name=table_1_save_name, update_mode=LoadUpdateMode.APPEND)
    publish_flow_test(
        client,
        existing_flow=flow,
        artifacts=table,
        engine=engine,
    )
    validator.check_saved_artifact(
        flow,
        table.id(),
        expected_data=pd.concat(
            [extracted_table_data, extracted_table_data, extracted_table_data], ignore_index=True
        ),
    )

    # Create a different table from the same artifact.
    save(integration, table, name=table_2_save_name, update_mode=LoadUpdateMode.REPLACE)
    publish_flow_test(
        client,
        existing_flow=flow,
        artifacts=table,
        engine=engine,
    )
    validator.check_saved_artifact(
        flow,
        table.id(),
        expected_data=extracted_table_data,
    )

    validator.check_saved_update_mode_changes(
        flow,
        expected_updates=[
            (table_2_save_name, LoadUpdateMode.REPLACE),
            (table_1_save_name, LoadUpdateMode.APPEND),
            (table_1_save_name, LoadUpdateMode.REPLACE),
        ],
    )


def test_multiple_artifacts_saved_to_same_integration(
    client, flow_name, data_integration, engine, validator
):
    integration = client.integration(name=data_integration)
    table_1_save_name = generate_table_name()
    table_2_save_name = generate_table_name()

    table_1 = integration.sql(query=SHORT_SENTIMENT_SQL_QUERY)
    save(integration, table_1, name=table_1_save_name, update_mode=LoadUpdateMode.REPLACE)
    table_2 = integration.sql(query=SHORT_SENTIMENT_SQL_QUERY)
    save(integration, table_2, name=table_2_save_name, update_mode=LoadUpdateMode.REPLACE)

    flow = publish_flow_test(
        client,
        name=flow_name(),
        artifacts=[table_1, table_2],
        engine=engine,
    )

    validator.check_saved_artifact(flow, table_1.id(), expected_data=table_1.get())
    validator.check_saved_artifact(flow, table_2.id(), expected_data=table_2.get())
    validator.check_saved_update_mode_changes(
        flow,
        expected_updates=[
            (table_2_save_name, LoadUpdateMode.REPLACE),
            (table_1_save_name, LoadUpdateMode.REPLACE),
        ],
        order_matters=False,
    )


def test_lazy_artifact_with_save(client, flow_name, data_integration, engine, validator):
    db = client.integration(data_integration)
    reviews = db.sql(SHORT_SENTIMENT_SQL_QUERY)

    @op()
    def copy_field(df):
        df["new"] = df["review"]
        return df

    review_copied = copy_field.lazy(reviews)
    save(db, review_copied, update_mode=LoadUpdateMode.REPLACE)

    flow = publish_flow_test(
        client,
        review_copied,
        name=flow_name(),
        engine=engine,
    )
    validator.check_saved_artifact(
        flow, review_copied.id(), expected_data=copy_field.local(reviews)
    )
