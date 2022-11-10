from os import cpu_count

from aqueduct import op
from utils import generate_new_flow_name, run_flow_test


# TODO: narrow this to only K8s
def test_custom_num_cpus(client, engine):
    """Assumption: nodes in the K8s cluster have more than 6 CPUs."""

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

    # Returns the default number of CPUs of the K8s cluster. (Currently 2)
    @op(requirements=[])
    def count_default_available_cpus():
        return _count_available_cpus()

    num_default_available_cpus = count_default_available_cpus.lazy()

    # Returns 6, the custom number of CPUs on the K8s cluster.
    @op(requirements=[], resources={"num_cpus": 6})
    def count_with_custom_available_cpus():
        return _count_available_cpus()

    num_count_available_cpus = count_with_custom_available_cpus.lazy()

    flows = []
    try:
        default_cpus_flow = run_flow_test(
            client,
            name=generate_new_flow_name(),
            artifacts=num_default_available_cpus,
            engine=engine,
            delete_flow_after=False,
        )
        flows.append(default_cpus_flow)

        custom_cpus_flow = run_flow_test(
            client,
            name=generate_new_flow_name(),
            artifacts=num_count_available_cpus,
            engine=engine,
            delete_flow_after=False,
        )
        flows.append(custom_cpus_flow)

        assert (
            default_cpus_flow.latest().artifact("count_default_available_cpus artifact").get() == 2
        )
        assert custom_cpus_flow.latest().artifact("count_with_custom_available_cpus artifact").get() == 6

    finally:
        for flow in flows:
            client.delete_flow(flow.id())
