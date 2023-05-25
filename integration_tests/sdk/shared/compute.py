from aqueduct.constants.enums import ServiceType


def type_from_engine_name(client, engine: str) -> ServiceType:
    """
    Returns the resource type of an engine from the name.
    """
    assert engine != "aqueduct_engine"

    if engine is None:
        return ServiceType.AQUEDUCT_ENGINE

    resource_info_by_name = client.list_resources()
    if engine not in resource_info_by_name.keys():
        raise Exception("Server is not connected to resource `%s`." % engine)

    return resource_info_by_name[engine].service
