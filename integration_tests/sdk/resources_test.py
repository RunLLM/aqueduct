from os import cpu_count

import pytest
from aqueduct.constants.enums import ExecutionStatus, ServiceType
from aqueduct.error import AqueductError, InvalidUserArgumentException
from utils import publish_flow_test

from aqueduct import global_config, op


@pytest.mark.enable_only_for_engine_type(ServiceType.K8S)
def test_custom_num_cpus(client, flow_name, engine):
    """Assumption: nodes in the K8s cluster have more than 4 CPUs.

    We run a special operator that checks the number of CPUs that are available.
    We check the expected default number of cpus, as well as a custom number.
    """

    def _count_available_cpus():
        # Copied from: https://donghao.org/2022/01/20/how-to-get-the-number-of-cpu-cores-inside-a-container/
        with open("/sys/fs/cgroup/cpu/cpu.cfs_quota_us") as fp:
            cfs_quota_us = int(fp.read())
        with open("/sys/fs/cgroup/cpu/cpu.cfs_period_us") as fp:
            cfs_period_us = int(fp.read())

        container_cpus = cfs_quota_us // cfs_period_us
        # For physical machine, the `cfs_quota_us` could be '-1'
        cpus = cpu_count() if container_cpus < 1 else container_cpus
        return cpus

    global_config({"engine": engine})

    # Returns the default number of CPUs of the K8s cluster. (Currently 2)
    @op(requirements=[])
    def count_default_available_cpus():
        return _count_available_cpus()

    num_default_available_cpus = count_default_available_cpus()
    assert num_default_available_cpus.get() == 2

    # Returns 4, the custom number of CPUs on the K8s cluster.
    @op(requirements=[], resources={"num_cpus": 4})
    def count_with_custom_available_cpus():
        return _count_available_cpus()

    num_count_available_cpus = count_with_custom_available_cpus()
    assert num_count_available_cpus.get() == 4

    default_cpus_flow = publish_flow_test(
        client,
        name=flow_name(),
        artifacts=num_default_available_cpus,
        engine=engine,
        delete_flow_after=False,
    )

    custom_cpus_flow = publish_flow_test(
        client,
        name=flow_name(),
        artifacts=num_count_available_cpus,
        engine=engine,
        delete_flow_after=False,
    )

    assert default_cpus_flow.latest().artifact("count_default_available_cpus artifact").get() == 2
    assert (
        custom_cpus_flow.latest().artifact("count_with_custom_available_cpus artifact").get() == 6
    )


@pytest.mark.enable_only_for_engine_type(ServiceType.K8S)
def test_too_many_cpus_requested(client, flow_name, engine):
    """Assumption: nodes in the k8s cluster have less then 20 CPUs."""
    global_config({"engine": engine})

    @op(requirements=[], resources={"num_cpus": 20})
    def too_many_cpus():
        return 123

    with pytest.raises(AqueductError):
        too_many_cpus()
    output = too_many_cpus.lazy()

    publish_flow_test(
        client, output, name=flow_name(), engine=engine, expected_statuses=ExecutionStatus.FAILED
    )


@pytest.mark.enable_only_for_engine_type(ServiceType.K8S, ServiceType.LAMBDA)
def test_custom_memory(client, flow_name, engine):
    """Assumption: nodes in the K8s cluster have more than 200MB of capacity.

    Customize our memory to be 200MB. We will run two different methods, one that allocates less than
    this amount and one that allocates more. The latter should fail.
    """
    global_config({"engine": engine})

    @op(requirements=[], resources={"memory": "200MB"})
    def fn_expect_success():
        return 123

    success_output = fn_expect_success()

    @op(requirements=[], resources={"memory": "200MB"})
    def fn_expect_failure():
        # Overallocate memory at runtime.
        output = bytearray(1000 * 1000 * 100 * 4)
        return output

    with pytest.raises(AqueductError):
        fn_expect_failure()
    failure_output = fn_expect_failure.lazy()

    publish_flow_test(
        client,
        name=flow_name(),
        artifacts=success_output,
        engine=engine,
    )

    publish_flow_test(
        client,
        name=flow_name,
        artifacts=failure_output,
        engine=engine,
        expected_statuses=ExecutionStatus.FAILED,
    )


@pytest.mark.enable_only_for_engine_type(ServiceType.K8S)
def test_too_much_memory_requested_K8s(client, flow_name, engine):
    """Assumption: nodes in the k8s cluster have less then 100GB of memory."""
    global_config({"engine": engine})

    @op(requirements=[], resources={"memory": "100GB"})
    def too_much_memory():
        return 123

    with pytest.raises(AqueductError):
        _ = too_much_memory()
    output = too_much_memory.lazy()

    publish_flow_test(
        client, [output], name=flow_name(), engine=engine, expected_statuses=ExecutionStatus.FAILED
    )


@pytest.mark.enable_only_for_engine_type(ServiceType.LAMBDA)
def test_too_much_memory_requested_lambda(client, flow_name, engine):
    """Lambda does not allow you to allocate more than 10,240MB of memory."""
    global_config({"engine": engine})

    @op(requirements=[], resources={"memory": 11000})
    def too_much_memory():
        return 123

    with pytest.raises(InvalidUserArgumentException):
        _ = too_much_memory()

    output = too_much_memory.lazy()

    with pytest.raises(InvalidUserArgumentException):
        publish_flow_test(client, name=flow_name(), artifacts=[output], engine=engine)


@pytest.mark.enable_only_for_engine_type(ServiceType.K8S)
def test_custom_gpus(client, flow_name, engine):
    """Assumption: there is a GPU node in the K8s cluster. Also assumes the
    machine executing the test has pytorch installed.

    We run a special operator that checks the availability of GPUs.
    """
    import torch

    global_config({"engine": engine})

    # Returns availability of GPU, should be False.
    @op(requirements=["torch==1.13.0"])
    def gpu_is_not_available():
        return torch.cuda.is_available()

    gpu_not_available = gpu_is_not_available()

    # Returns availability of GPU, should be True.
    @op(requirements=["torch==1.13.0"], resources={"gpu_resource_name": "nvidia.com/gpu"})
    def gpu_is_available():
        return torch.cuda.is_available()

    gpu_available = gpu_is_available()

    no_gpu_flow = publish_flow_test(
        client,
        name=flow_name(),
        artifacts=gpu_not_available,
        engine=engine,
    )

    gpu_flow = publish_flow_test(
        client,
        name=flow_name(),
        artifacts=gpu_available,
        engine=engine,
    )

    assert not no_gpu_flow.latest().artifact("gpu_is_not_available artifact").get()
    assert gpu_flow.latest().artifact("gpu_is_available artifact").get()
