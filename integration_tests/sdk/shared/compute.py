from aqueduct.constants.enums import ServiceType


def type_from_engine_name(client, engine: str) -> ServiceType:
    """
    Returns the integration type of an engine from the name.
    """
    if engine == "aqueduct_engine":
        return ServiceType.AQUEDUCT_ENGINE

    integration_info_by_name = client.list_integrations()
    if engine not in integration_info_by_name.keys():
        raise Exception("Server is not connected to integration `%s`." % engine)

    return integration_info_by_name[engine].service
