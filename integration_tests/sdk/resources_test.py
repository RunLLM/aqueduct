

def test_cpu_resource_constraint(client, engine):
    @op(requirements=[])