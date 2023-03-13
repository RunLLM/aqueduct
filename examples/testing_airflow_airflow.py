from datetime import datetime

from airflow.models import DAG
from airflow.operators.python import PythonVirtualenvOperator

# Python requirements for each Airflow task
VENV_REQUIREMENTS=[
    "aqueduct-ml",
    "awswrangler",
    "boto3",
    "cloudpickle",
    "distro",
    "google-cloud-bigquery",
    "mysqlclient",
    "pandas",
    "psycopg2-binary",
    "pyarrow",
    "pydantic",
    "pyodbc",
    "pyyaml",
    "scikit-learn",
    "snowflake-sqlalchemy",
    "SQLAlchemy",
    "typing_extensions",
]

def invoke_task(spec, **kwargs):
    '''
    Check the spec type and invoke the correct operator.
    First, append the dag_run_id to all of the storage paths in the spec.
    '''
    from aqueduct_executor.operators.utils import enums
    from aqueduct_executor.operators.function_executor import spec as func_spec
    from aqueduct_executor.operators.param_executor import spec as param_spec
    from aqueduct_executor.operators.system_metric_executor import spec as sys_metric_spec
    from aqueduct_executor.operators.connectors.data import spec as conn_spec

    from aqueduct_executor.operators.function_executor import execute as func_execute
    from aqueduct_executor.operators.param_executor import execute as param_execute
    from aqueduct_executor.operators.system_metric_executor import execute as sys_metric_execute
    from aqueduct_executor.operators.connectors.data import execute as conn_execute

    dag_run_id = kwargs["run_id"]

    spec_type = enums.JobType(spec["type"])
    if spec_type == enums.JobType.FUNCTION:
        spec = func_spec.FunctionSpec(**spec)
        spec.metadata_path = "{}_{}".format(spec.metadata_path, dag_run_id)
        spec.input_content_paths = ["{}_{}".format(p, dag_run_id) for p in spec.input_content_paths]
        spec.input_metadata_paths = ["{}_{}".format(p, dag_run_id) for p in spec.input_metadata_paths]
        spec.output_content_paths = ["{}_{}".format(p, dag_run_id) for p in spec.output_content_paths]
        spec.output_metadata_paths = ["{}_{}".format(p, dag_run_id) for p in spec.output_metadata_paths]
        func_execute.run_with_setup(spec)
    elif spec_type == enums.JobType.EXTRACT:
        spec = conn_spec.ExtractSpec(**spec)
        spec.metadata_path = "{}_{}".format(spec.metadata_path, dag_run_id)
        spec.output_content_path = "{}_{}".format(spec.output_content_path, dag_run_id)
        spec.output_metadata_path = "{}_{}".format(spec.output_metadata_path, dag_run_id)
        conn_execute.run(spec)
    elif spec_type == enums.JobType.LOAD:
        spec = conn_spec.LoadSpec(**spec)
        spec.metadata_path = "{}_{}".format(spec.metadata_path, dag_run_id)
        spec.input_content_path = "{}_{}".format(spec.input_content_path, dag_run_id)
        spec.input_metadata_path = "{}_{}".format(spec.input_metadata_path, dag_run_id)
        conn_execute.run(spec)
    elif spec_type == enums.JobType.PARAM:
        spec = param_spec.ParamSpec(**spec)
        spec.metadata_path = "{}_{}".format(spec.metadata_path, dag_run_id)
        spec.output_content_path = "{}_{}".format(spec.output_content_path, dag_run_id)
        spec.output_metadata_path = "{}_{}".format(spec.output_metadata_path, dag_run_id)
        param_execute.run(spec)
    elif spec_type == enums.JobType.SYSTEM_METRIC:
        spec = sys_metric_spec.SystemMetricSpec(**spec)
        spec.metadata_path = "{}_{}".format(spec.metadata_path, dag_run_id)
        spec.input_metadata_paths = ["{}_{}".format(p, dag_run_id) for p in spec.input_metadata_paths]
        spec.output_content_path = "{}_{}".format(spec.output_content_path, dag_run_id)
        spec.output_metadata_path = "{}_{}".format(spec.output_metadata_path, dag_run_id)
        sys_metric_execute.run(spec)


with DAG(
    dag_id='testing_airflow',
    default_args={
        'retries': 0,
    },
    start_date=datetime(2022, 1, 1, 1),
    
    schedule_interval=None,
    
    catchup=False,
    tags=['aqueduct', 'b8252427-7fa9-4478-a3cc-6b858d2bb800'],
) as dag:
    # Constants to handle JSON serialization
    null = None
    false = False
    true = True


    
    t0 = PythonVirtualenvOperator(
        task_id='noop',
        requirements=VENV_REQUIREMENTS,
        system_site_packages=False,
        python_callable=invoke_task,
        op_args=[
    {
    "name": "function-operator-d64d136c-e12c-4db0-a0af-1bd5dc88f000",
    "type": "function",
    "storage_config": {
        "type": "s3",
        "file_config": null,
        "s3_config": {
            "region": "us-east-2",
            "bucket": "s3://sauravoss",
            "credentials_path": "/home/ubuntu/.aws/credentials",
            "credentials_profile": "default",
            "aws_access_key_id": "",
            "aws_secret_access_key": ""
        },
        "gcs_config": null
    },
    "metadata_path": "2830225f-9958-42dd-9dcb-53c41657bb4d",
    "function_path": "operator-63b799f5-42b7-4bff-a055-8c1c2b49b736",
    "function_extract_path": "",
    "entry_point_file": "model.py",
    "entry_point_class": "Function",
    "entry_point_method": "predict",
    "custom_args": "",
    "input_content_paths": [
        "29728854-d32c-4d8d-aaa1-ca1628b5e2b2"
    ],
    "input_metadata_paths": [
        "ccbc81b9-b3eb-44ba-90f9-ecd7d9f2096f"
    ],
    "output_content_paths": [
        "4912050f-7e85-4d57-b092-396fd7662cb8"
    ],
    "output_metadata_paths": [
        "9700ad49-c60f-47f4-8487-3cd2601d3f48"
    ],
    "expected_output_artifact_types": [
        "table"
    ],
    "operator_type": "function",
    "check_severity": null,
    "resources": null
}
        ],
    )
    
    t1 = PythonVirtualenvOperator(
        task_id='sflake_query',
        requirements=VENV_REQUIREMENTS,
        system_site_packages=False,
        python_callable=invoke_task,
        op_args=[
    {
    "name": "extract-operator-74af1a4c-eeed-4281-b190-de19dc35dc1d",
    "type": "extract",
    "storage_config": {
        "type": "s3",
        "file_config": null,
        "s3_config": {
            "region": "us-east-2",
            "bucket": "s3://sauravoss",
            "credentials_path": "/home/ubuntu/.aws/credentials",
            "credentials_profile": "default",
            "aws_access_key_id": "",
            "aws_secret_access_key": ""
        },
        "gcs_config": null
    },
    "metadata_path": "d1f09a37-025b-4590-8822-3d3662d40b20",
    "connector_name": "Snowflake",
    "connector_config": {
        "username": "SAURAV",
        "password": "djNoalx38$$jsQTeO5XLt[3Hwy",
        "account_identifier": "baa81868",
        "database": "SAURAV",
        "warehouse": "COMPUTE_WH",
        "db_schema": "public",
        "role": null
    },
    "parameters": {
        "query_is_usable": false,
        "query": "select * from wine;",
        "queries": null,
        "github_metadata": null
    },
    "input_param_names": [],
    "input_content_paths": [],
    "input_metadata_paths": [],
    "output_content_path": "29728854-d32c-4d8d-aaa1-ca1628b5e2b2",
    "output_metadata_path": "ccbc81b9-b3eb-44ba-90f9-ecd7d9f2096f"
}
        ],
    )
    


    t1.set_downstream(t0)
