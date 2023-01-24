import uuid


def generate_new_flow_name() -> str:
    return "test_" + uuid.uuid4().hex


def generate_table_name() -> str:
    return "test_table_" + uuid.uuid4().hex[:24]


def generate_object_name() -> str:
    """For non-relational data integrations."""
    return "test_object_" + uuid.uuid4().hex[:24]
