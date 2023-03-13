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
    dag_id='Wine_Ratings',
    default_args={
        'retries': 0,
    },
    start_date=datetime(2022, 1, 1, 1),
    
    schedule_interval='0 0 * * *',
    
    catchup=False,
    tags=['aqueduct', '2e1e15bf-d683-424b-8544-bde9903e3f95'],
) as dag:
    # Constants to handle JSON serialization
    null = None
    false = False
    true = True


    
    t0 = PythonVirtualenvOperator(
        task_id='encode_color',
        requirements=VENV_REQUIREMENTS,
        system_site_packages=False,
        python_callable=invoke_task,
        op_args=[
    {
    "name": "function-operator-60ec1388-1193-4b59-a9a2-8737c6a59000",
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
    "metadata_path": "10036e21-a889-4275-a71c-1f4edd1a830f",
    "function_path": "operator-724f6ca3-c350-49a3-a466-205e3162194f",
    "function_extract_path": "",
    "entry_point_file": "model.py",
    "entry_point_class": "Function",
    "entry_point_method": "predict",
    "custom_args": "",
    "input_content_paths": [
        "1ff84015-f35d-46ef-a5f0-a6af8139d35e"
    ],
    "input_metadata_paths": [
        "1c2828e0-f162-4af4-bc2b-e7df01d1e1e6"
    ],
    "output_content_paths": [
        "576c2346-7d77-47a6-986c-f53db534bc43"
    ],
    "output_metadata_paths": [
        "9a812c08-9bac-4ec6-9343-11dc61c60e39"
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
        task_id='fix_residual_sugar',
        requirements=VENV_REQUIREMENTS,
        system_site_packages=False,
        python_callable=invoke_task,
        op_args=[
    {
    "name": "function-operator-5bd9d83a-cdef-4384-abd3-1c8299b0924b",
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
    "metadata_path": "f39ce0fb-0935-4d70-a0ae-1a8a4564ba70",
    "function_path": "operator-e7514fac-3baa-47f6-8c9a-ac09b3e1cfed",
    "function_extract_path": "",
    "entry_point_file": "model.py",
    "entry_point_class": "Function",
    "entry_point_method": "predict",
    "custom_args": "",
    "input_content_paths": [
        "7669a7df-a8ae-490a-8fab-953511f76295"
    ],
    "input_metadata_paths": [
        "507c309c-7270-44c7-af26-9ff8a54a4590"
    ],
    "output_content_paths": [
        "1ff84015-f35d-46ef-a5f0-a6af8139d35e"
    ],
    "output_metadata_paths": [
        "1c2828e0-f162-4af4-bc2b-e7df01d1e1e6"
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
    
    t2 = PythonVirtualenvOperator(
        task_id='get_number_labeled_wines',
        requirements=VENV_REQUIREMENTS,
        system_site_packages=False,
        python_callable=invoke_task,
        op_args=[
    {
    "name": "function-operator-efc3d8d8-e812-4436-8721-ec871919aec5",
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
    "metadata_path": "3268d2f5-d5cc-4252-9a87-ca456014c6d6",
    "function_path": "operator-628015c3-5409-45ba-afe0-09e44b647694",
    "function_extract_path": "",
    "entry_point_file": "model.py",
    "entry_point_class": "Function",
    "entry_point_method": "predict",
    "custom_args": "",
    "input_content_paths": [
        "576c2346-7d77-47a6-986c-f53db534bc43"
    ],
    "input_metadata_paths": [
        "9a812c08-9bac-4ec6-9343-11dc61c60e39"
    ],
    "output_content_paths": [
        "39b72d4f-1c4d-448c-959e-41e4c497c1cb"
    ],
    "output_metadata_paths": [
        "2cdd1e56-907d-43c6-b200-35c248d41284"
    ],
    "expected_output_artifact_types": [
        "numeric"
    ],
    "operator_type": "metric",
    "check_severity": null,
    "resources": null
}
        ],
    )
    
    t3 = PythonVirtualenvOperator(
        task_id='get_rmse',
        requirements=VENV_REQUIREMENTS,
        system_site_packages=False,
        python_callable=invoke_task,
        op_args=[
    {
    "name": "function-operator-2f9ef1c1-ce32-4b06-b0df-784e84e91e98",
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
    "metadata_path": "1f5a4ba1-8730-41cb-b0a3-4457adf007f2",
    "function_path": "operator-b2f04d7f-72bd-4f56-b5ff-fb2bf81607f8",
    "function_extract_path": "",
    "entry_point_file": "model.py",
    "entry_point_class": "Function",
    "entry_point_method": "predict",
    "custom_args": "",
    "input_content_paths": [
        "e9881044-9470-4651-81b3-a101e02462c8"
    ],
    "input_metadata_paths": [
        "0060feed-73fe-4684-8a0c-364b74bcc685"
    ],
    "output_content_paths": [
        "88cec62c-3e28-4311-b901-985647714cf1"
    ],
    "output_metadata_paths": [
        "a4ca1227-cf41-48ab-b453-c846ced4f306"
    ],
    "expected_output_artifact_types": [
        "numeric"
    ],
    "operator_type": "metric",
    "check_severity": null,
    "resources": null
}
        ],
    )
    
    t4 = PythonVirtualenvOperator(
        task_id='greater_than_1000',
        requirements=VENV_REQUIREMENTS,
        system_site_packages=False,
        python_callable=invoke_task,
        op_args=[
    {
    "name": "function-operator-fb7623f2-36c5-4e48-bed0-af07a2ace1bb",
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
    "metadata_path": "db9ed83b-e09e-4538-b6ae-83dc205e259a",
    "function_path": "operator-5d32dfee-76c7-43c0-9bd3-a71d08de9adf",
    "function_extract_path": "",
    "entry_point_file": "model.py",
    "entry_point_class": "Function",
    "entry_point_method": "predict",
    "custom_args": "",
    "input_content_paths": [
        "39b72d4f-1c4d-448c-959e-41e4c497c1cb"
    ],
    "input_metadata_paths": [
        "2cdd1e56-907d-43c6-b200-35c248d41284"
    ],
    "output_content_paths": [
        "64c93780-6ac7-485a-9e29-ea1af326aa8e"
    ],
    "output_metadata_paths": [
        "278fa0ad-f9a7-4f1f-bdd3-dd4850756b90"
    ],
    "expected_output_artifact_types": [
        "boolean"
    ],
    "operator_type": "check",
    "check_severity": "error",
    "resources": null
}
        ],
    )
    
    t5 = PythonVirtualenvOperator(
        task_id='less_than_1_0',
        requirements=VENV_REQUIREMENTS,
        system_site_packages=False,
        python_callable=invoke_task,
        op_args=[
    {
    "name": "function-operator-43585ab1-8314-43d3-b1b7-a1b722902353",
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
    "metadata_path": "551176e8-7af7-4aa3-b57e-171bcb43db4c",
    "function_path": "operator-6350d0c5-a2a4-473b-bb7a-36d05de50493",
    "function_extract_path": "",
    "entry_point_file": "model.py",
    "entry_point_class": "Function",
    "entry_point_method": "predict",
    "custom_args": "",
    "input_content_paths": [
        "88cec62c-3e28-4311-b901-985647714cf1"
    ],
    "input_metadata_paths": [
        "a4ca1227-cf41-48ab-b453-c846ced4f306"
    ],
    "output_content_paths": [
        "81e438b0-5d0a-4566-973b-a1c00b79c88a"
    ],
    "output_metadata_paths": [
        "61ffe63c-431a-4a1f-b93a-03e85dd64510"
    ],
    "expected_output_artifact_types": [
        "boolean"
    ],
    "operator_type": "check",
    "check_severity": "warning",
    "resources": null
}
        ],
    )
    
    t6 = PythonVirtualenvOperator(
        task_id='less_than_3_0',
        requirements=VENV_REQUIREMENTS,
        system_site_packages=False,
        python_callable=invoke_task,
        op_args=[
    {
    "name": "function-operator-8aa1c173-f094-4724-b4e2-0a8941b31b6e",
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
    "metadata_path": "63fa3387-0b88-4d83-8b29-0e7e4fcee4ab",
    "function_path": "operator-8cbc05fd-7740-41aa-8299-cbe2f19aa5c6",
    "function_extract_path": "",
    "entry_point_file": "model.py",
    "entry_point_class": "Function",
    "entry_point_method": "predict",
    "custom_args": "",
    "input_content_paths": [
        "88cec62c-3e28-4311-b901-985647714cf1"
    ],
    "input_metadata_paths": [
        "a4ca1227-cf41-48ab-b453-c846ced4f306"
    ],
    "output_content_paths": [
        "d5047409-d58c-4ee9-993d-d15a20b880ae"
    ],
    "output_metadata_paths": [
        "80dc4fcf-4072-4f35-a33f-c6c8900e4a57"
    ],
    "expected_output_artifact_types": [
        "boolean"
    ],
    "operator_type": "check",
    "check_severity": "error",
    "resources": null
}
        ],
    )
    
    t7 = PythonVirtualenvOperator(
        task_id='predict_quality',
        requirements=VENV_REQUIREMENTS,
        system_site_packages=False,
        python_callable=invoke_task,
        op_args=[
    {
    "name": "function-operator-46f64541-1d8e-41eb-a730-59638a676f8d",
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
    "metadata_path": "05b25d88-dcb4-4bf5-a352-efa61f1b4197",
    "function_path": "operator-aa2b0df5-2aa7-41ad-b1c6-f8b4fa0b20ee",
    "function_extract_path": "",
    "entry_point_file": "model.py",
    "entry_point_class": "Function",
    "entry_point_method": "predict",
    "custom_args": "",
    "input_content_paths": [
        "576c2346-7d77-47a6-986c-f53db534bc43"
    ],
    "input_metadata_paths": [
        "9a812c08-9bac-4ec6-9343-11dc61c60e39"
    ],
    "output_content_paths": [
        "e9881044-9470-4651-81b3-a101e02462c8"
    ],
    "output_metadata_paths": [
        "0060feed-73fe-4684-8a0c-364b74bcc685"
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
    
    t8 = PythonVirtualenvOperator(
        task_id='save_to_sflake',
        requirements=VENV_REQUIREMENTS,
        system_site_packages=False,
        python_callable=invoke_task,
        op_args=[
    {
    "name": "load-operator-568d85b5-c1ec-42e1-80bc-d89c04075514",
    "type": "load",
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
    "metadata_path": "f2b3a140-e5ba-4409-ae08-fda698d5e5b0",
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
        "table": "pred_wine_quality",
        "update_mode": "replace"
    },
    "input_content_path": "e9881044-9470-4651-81b3-a101e02462c8",
    "input_metadata_path": "0060feed-73fe-4684-8a0c-364b74bcc685"
}
        ],
    )
    
    t9 = PythonVirtualenvOperator(
        task_id='sflake_query',
        requirements=VENV_REQUIREMENTS,
        system_site_packages=False,
        python_callable=invoke_task,
        op_args=[
    {
    "name": "extract-operator-cd828dfc-d67c-492a-bf0f-268520fa4d08",
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
    "metadata_path": "eadc58c0-7ce7-4a26-8471-0ee939187c43",
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
    "output_content_path": "7669a7df-a8ae-490a-8fab-953511f76295",
    "output_metadata_path": "507c309c-7270-44c7-af26-9ff8a54a4590"
}
        ],
    )
    


    t0.set_downstream(t7)

    t0.set_downstream(t2)

    t1.set_downstream(t0)

    t2.set_downstream(t4)

    t3.set_downstream(t5)

    t3.set_downstream(t6)

    t7.set_downstream(t3)

    t7.set_downstream(t8)

    t9.set_downstream(t1)
