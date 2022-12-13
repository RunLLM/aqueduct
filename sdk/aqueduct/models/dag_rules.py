from typing import Dict

from aqueduct.constants.enums import RuntimeType

from ..error import InvalidUserArgumentException
from ..logger import logger
from .config import EngineConfig
from .operators import LAMBDA_MAX_MEMORY_MB, LAMBDA_MIN_MEMORY_MB, ResourceConfig


def check_customized_resources_are_supported(
    resources: ResourceConfig,
    engine_config: EngineConfig,
    op_name: str,
) -> None:
    allowed_customizable_resources: Dict[str, bool] = {
        "num_cpus": False,
        "memory": False,
        "gpu_resource_name": False,
    }
    if engine_config.type == RuntimeType.K8S:
        allowed_customizable_resources = {
            "num_cpus": True,
            "memory": True,
            "gpu_resource_name": True,
        }
    elif engine_config.type == RuntimeType.LAMBDA:
        allowed_customizable_resources["memory"] = True

    if not allowed_customizable_resources["num_cpus"] and resources.num_cpus:
        raise InvalidUserArgumentException(
            "Operator `%s` cannot configure the number of cpus, since it is not supported when running on %s."
            % (op_name, engine_config.type)
        )

    if not allowed_customizable_resources["memory"] and resources.memory_mb:
        raise InvalidUserArgumentException(
            "Operator `%s` cannot configure the amount of memory, since it is not supported when running on %s."
            % (op_name, engine_config.type)
        )

    if engine_config.type == RuntimeType.LAMBDA and resources.memory_mb:
        if resources.memory_mb < LAMBDA_MIN_MEMORY_MB:
            raise InvalidUserArgumentException(
                "AWS Lambda method must be configured with at least %d MB of memory, but got request for %d."
                % (LAMBDA_MIN_MEMORY_MB, resources.memory_mb)
            )
        elif resources.memory_mb > LAMBDA_MAX_MEMORY_MB:
            raise InvalidUserArgumentException(
                "AWS Lambda method must be configured with at most %d MB of memory, but got a request for %d."
                % (LAMBDA_MIN_MEMORY_MB, resources.memory_mb)
            )
        logger().warning(
            "Customizing memory for a AWS Lambda operator will add about a minute to its runtime, per operator."
        )

    if not allowed_customizable_resources["gpu_resource_name"] and resources.gpu_resource_name:
        raise InvalidUserArgumentException(
            "Operator `%s` cannot configure gpus, since it is not supported when running on %s."
            % (op_name, engine_config.type)
        )
