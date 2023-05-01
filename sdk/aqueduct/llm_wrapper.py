from typing import Any, Callable, Dict, List, Optional, Union

import pandas as pd
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.constants.enums import RuntimeType
from aqueduct.decorator import op
from aqueduct.error import InvalidUserArgumentException
from aqueduct.utils.utils import generate_engine_config

from aqueduct import globals

supported_llm_models = [
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


def generate_llm_op(
    llm_name: str, column_name: Optional[str] = None, output_column_name: Optional[str] = None
) -> Callable[
    [Union[str, List[str], pd.DataFrame], Dict[str, Any]], Union[str, List[str], pd.DataFrame]
]:
    def use_llm(
        messages: Union[str, List[str]], parameters: Dict[str, Any] = {}
    ) -> Union[str, List[str]]:
        if not (isinstance(messages, str) or isinstance(messages, list)):
            raise ValueError("The 'messages' parameter must be a string or list of strings.")

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
            raise ValueError("The 'df' parameter must be a pandas DataFrame.")

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
    llm_name: str,
    op_name: Optional[str] = None,
    engine: Optional[str] = None,
    column_name: Optional[str] = None,
    output_column_name: Optional[str] = None,
) -> Union[
    Callable[..., Union[BaseArtifact, List[BaseArtifact]]], BaseArtifact, List[BaseArtifact]
]:
    if llm_name not in supported_llm_models:
        raise InvalidUserArgumentException(f"Unsupported LLM model {llm_name}")

    kwargs: Dict[str, Any] = {}
    if engine is not None:
        kwargs["engine"] = engine

        engine_config = generate_engine_config(
            globals.__GLOBAL_API_CLIENT__.list_integrations(),
            engine,
        )
        if engine_config and engine_config.type == RuntimeType.K8S:
            kwargs["resources"] = resource_requests[llm_name]

    if op_name is None:
        op_name = llm_name

    return op(
        name=op_name,
        requirements=["aqueduct-llm"],
        **kwargs,
    )(
        generate_llm_op(
            llm_name=llm_name, column_name=column_name, output_column_name=output_column_name
        )
    )
