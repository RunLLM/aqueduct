from typing import Any, Callable, Dict, List, Optional, Union

import pandas as pd
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.constants.enums import RuntimeType
from aqueduct.decorator import op
from aqueduct.error import InvalidUserArgumentException
from aqueduct.resources.dynamic_k8s import DynamicK8sResource
from aqueduct.utils.utils import generate_engine_config

from aqueduct import globals

supported_llms = [
    "llama_7b",
    "vicuna_7b",
    "dolly_v2_3b",
    "dolly_v2_7b",
]

resource_requests = {
    "llama_7b": {
        "memory": "16GB",
        "gpu_resource_name": "nvidia.com/gpu",
    },
    "vicuna_7b": {
        "memory": "32GB",
        "gpu_resource_name": "nvidia.com/gpu",
    },
    "dolly_v2_3b": {
        "memory": "8GB",
        "gpu_resource_name": "nvidia.com/gpu",
    },
    "dolly_v2_7b": {
        "memory": "16GB",
        "gpu_resource_name": "nvidia.com/gpu",
    },
}


def _generate_llm_op(
    llm_name: str, column_name: Optional[str] = None, output_column_name: Optional[str] = None
) -> Callable[
    [Union[str, List[str], pd.DataFrame], Dict[str, Any]], Union[str, List[str], pd.DataFrame]
]:
    def use_llm(
        messages: Union[str, List[str]], parameters: Dict[str, Any] = {}
    ) -> Union[str, List[str]]:
        if not (isinstance(messages, str) or isinstance(messages, list)):
            raise ValueError("The input must be a string or list of strings.")

        module = __import__("aqueduct_llm", fromlist=[llm_name])
        llm = getattr(module, llm_name)

        if "prompt" in parameters:
            prompt = parameters["prompt"]
            if not isinstance(prompt, str):
                raise ValueError("The 'prompt' parameter must be a string.")

            if "{text}" not in prompt:
                messages = (
                    prompt + " " + messages
                    if isinstance(messages, str)
                    else [prompt + " " + m for m in messages]
                )
            else:
                messages = (
                    prompt.replace("{text}", messages)
                    if isinstance(messages, str)
                    else [prompt.replace("{text}", m) for m in messages]
                )

            del parameters["prompt"]

        response = llm.generate(messages, **parameters)
        assert isinstance(response, str) or isinstance(response, list)
        return response

    def use_llm_for_table(df: pd.DataFrame, parameters: Dict[str, Any] = {}) -> pd.DataFrame:
        if not isinstance(df, pd.DataFrame):
            raise ValueError("The input must be a pandas DataFrame.")

        module = __import__("aqueduct_llm", fromlist=[llm_name])
        llm = getattr(module, llm_name)

        if "prompt" in parameters:
            prompt = parameters["prompt"]
            if not isinstance(prompt, str):
                raise ValueError("The 'prompt' parameter must be a string.")

            if "{text}" not in prompt:
                input_series = prompt + " " + df[column_name].astype(str)
            else:
                input_series = (
                    df[column_name].astype(str).apply(lambda x: prompt.replace("{text}", x))
                )

            del parameters["prompt"]
        else:
            input_series = df[column_name].astype(str)

        response = llm.generate(input_series.tolist(), **parameters)
        assert isinstance(response, list)

        df[output_column_name] = response
        return df

    if column_name is None and output_column_name is None:
        return use_llm
    else:
        if column_name is None or output_column_name is None:
            raise InvalidUserArgumentException(
                "Both column_name and output_column_name must be provided."
            )
        if not isinstance(column_name, str) or not isinstance(output_column_name, str):
            raise InvalidUserArgumentException(
                "column_name and output_column_name must be strings."
            )

        return use_llm_for_table


def llm_op(
    name: str,
    op_name: Optional[str] = None,
    column_name: Optional[str] = None,
    output_column_name: Optional[str] = None,
    engine: Optional[Union[str, DynamicK8sResource]] = None,
) -> Union[
    Callable[..., Union[BaseArtifact, List[BaseArtifact]]], BaseArtifact, List[BaseArtifact]
]:
    """Generates an Aqueduct operator to run a LLM. Either both column_name and output_column_name must be provided,
    or neither must be provided. Please refer to the `Returns` section below for their differences.
    Args:
        name:
            The name of the LLM to use. Please see aqueduct.supported_llms for a list of supported LLMs.
        op_name:
            The name of the operator. If not provided, defaults to the name of the LLM.
        column_name:
            The name of the column of the Dataframe to use as input to the LLM. If this field is provided,
            output_column_name must also be provided.
        output_column_name:
            The name of the column of the Dataframe to store the output of the LLM. If this field is provided,
            column_name must also be provided.

        engine:
            The name of the compute resource this operator will run on. Defaults to the Aqueduct engine.
            We recommend using a Kubernetes engine to run LLM operators, as we have implemented performance
            optimizations for LLMs on Kubernetes.
    Returns:
        If column_name and output_column_name are both provided, returns a function that takes in a
        DataFrame and returns a DataFrame with the output of the LLM appended as a new column:
        ```python
        def use_llm_for_table(df: pd.DataFrame, parameters: Dict[str, Any] = {}) -> pd.DataFrame:
        ```
        Otherwise, returns a function that takes in a string or list of strings, applies LLM, and
        returns a string or list of strings:
        ```python
        def use_llm(messages: Union[str, List[str]], parameters: Dict[str, Any] = {}) -> Union[str, List[str]]:
        ```
        In both cases, the function takes in an optional second argument, which is a dictionary of
        parameters to pass to the LLM. Please refer to the documentation for the LLM you are using
        for a list of supported parameters. For all LLMs, we support the "prompt" parameter. If the
        prompt contains {text}, we will replace {text} with the input string(s) before sending to
        the LLM. If the prompt does not contain {text}, we will prepend the prompt to the input
        string(s) before sending to the LLM.
    Examples:
        ```python
        >>> from aqueduct import Client
        >>> client = Client()
        >>> snowflake = client.resource("snowflake")
        >>> reviews_table = snowflake.sql("select * from hotel_reviews;")
        >>> from aqueduct import llm_op
        >>> vicuna_table_op = llm_op(
        ...     name="vicuna_7b",
        ...     op_name="my_vicuna_operator",
        ...     column_name="review",
        ...     output_column_name="response",
        ...     engine=ondemand_k8s,
        ... )
        >>> params = client.create_param(
        ...     "vicuna_params",
        ...     default={
        ...         "prompt": "Respond to the following hotel review as a customer service agent: {text} ",
        ...         "max_gpu_memory": "13GiB",
        ...         "temperature": 0.7,
        ...         "max_new_tokens": 512,
        ...     }
        ... )
        >>> review_with_response = vicuna_table_op(reviews_table, params)
        `review_with_response` is a Table Artifact with the output of the LLM appended as a new column.
        >>> review_with_response.get()
        ```
    """
    if name not in supported_llms:
        raise InvalidUserArgumentException(f"Unsupported LLM model {name}")

    kwargs: Dict[str, Any] = {}
    if engine is not None:
        kwargs["engine"] = engine

        engine_config = generate_engine_config(
            globals.__GLOBAL_API_CLIENT__.list_resources(),
            engine,
        )
        if engine_config and engine_config.type == RuntimeType.K8S:
            kwargs["resources"] = resource_requests[name]

    if op_name is None:
        op_name = name

    return op(
        name=op_name,
        requirements=["aqueduct-llm"],
        **kwargs,
    )(
        _generate_llm_op(
            llm_name=name, column_name=column_name, output_column_name=output_column_name
        )
    )
