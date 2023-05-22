from aqueduct.resources import GCSConfig, S3Config


def test_consistent_use_as_storage_field_name_for_storage_layer_resources():
    # Checks that all data connection configs have the same field name for using as storage layer.
    # This is a necessary assumption to enforce for our resource test setup.
    field_name = "use_as_storage"
    assert field_name in S3Config.__fields__.keys()
    assert field_name in GCSConfig.__fields__.keys()
