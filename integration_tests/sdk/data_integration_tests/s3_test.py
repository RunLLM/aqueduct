from typing import Optional

import pandas as pd
import pytest
from aqueduct.artifacts.base_artifact import BaseArtifact
from aqueduct.constants.enums import ArtifactType
from aqueduct.error import AqueductError, InvalidUserArgumentException
from aqueduct.flow import Flow
from aqueduct.resources.s3 import S3Resource
from PIL import Image

from aqueduct import op
from sdk.data_integration_tests.flow_manager import FlowManager
from sdk.data_integration_tests.s3_data_validator import S3DataValidator
from sdk.data_integration_tests.save import save
from sdk.data_integration_tests.validation_helpers import (
    check_hotel_reviews_table_artifact,
    check_hotel_reviews_table_data,
)
from sdk.shared.globals import artifact_id_to_saved_identifier
from sdk.shared.naming import generate_object_name, generate_table_name
from sdk.shared.validation import check_artifact_was_computed


@pytest.fixture(autouse=True)
def assert_data_integration_is_s3(data_integration):
    assert isinstance(data_integration, S3Resource)


def _save_artifact_and_check(
    flow_manager: FlowManager,
    data_integration: S3Resource,
    artifact: BaseArtifact,
    format: Optional[str],
    object_identifier: Optional[str] = None,
    skip_data_check: bool = False,
) -> Flow:
    """Saves the artifact by publishing a flow, and then checks that the data now exists in S3."""
    assert isinstance(artifact, BaseArtifact)

    if object_identifier is None:
        object_identifier = generate_table_name() if format is not None else generate_object_name()
    save(data_integration, artifact, object_identifier, format)

    flow = flow_manager.publish_flow_test(artifact)

    S3DataValidator(flow_manager._client, data_integration).check_saved_artifact_data(
        flow,
        artifact.id(),
        artifact.type(),
        format,
        expected_data=artifact.get(),
        skip_data_check=skip_data_check,
    )

    return flow


def test_s3_table_fetch_and_save(flow_manager, data_integration):
    hotel_reviews = data_integration.file(
        "hotel_reviews", artifact_type=ArtifactType.TABLE, format="parquet"
    )
    check_hotel_reviews_table_artifact(hotel_reviews)
    _save_artifact_and_check(flow_manager, data_integration, artifact=hotel_reviews, format="csv")


def test_s3_list_of_tables_fetch_and_save(client, flow_manager, data_integration):
    hotel_reviews = data_integration.file(
        "hotel_reviews", artifact_type=ArtifactType.TABLE, format="parquet"
    )

    @op
    def create_list_of_tables(input_table):
        return [input_table, input_table, input_table]

    list_output = create_list_of_tables(hotel_reviews)
    name = generate_object_name()
    _save_artifact_and_check(
        flow_manager,
        data_integration,
        artifact=list_output,
        format=None,
        object_identifier=name,
    )

    extracted_list_of_tables = data_integration.file(
        name,
        ArtifactType.LIST,
    )
    assert all(elem.equals(hotel_reviews.get()) for elem in extracted_list_of_tables.get())


def test_s3_custom_pickled_dictionaries_fetch_and_save(client, flow_manager, data_integration):
    # Current working directory is one level above.
    image_data = Image.open("data/aqueduct.jpg", "r")

    img_tuple_param = client.create_param("Image List", default=(image_data, image_data))
    img_tuple_identifier = generate_object_name()

    @op
    def return_image_list():
        return [image_data, image_data, image_data]

    img_list_artifact = return_image_list()
    img_list_identifier = generate_object_name()

    _save_artifact_and_check(
        flow_manager,
        data_integration,
        artifact=img_tuple_param,
        format=None,
        object_identifier=img_tuple_identifier,
        skip_data_check=True,
    )
    _save_artifact_and_check(
        flow_manager,
        data_integration,
        artifact=img_list_artifact,
        format=None,
        object_identifier=img_list_identifier,
        skip_data_check=True,
    )

    # Check that can smoothly extract these types of objects back.
    extracted_img_tuple = data_integration.file(
        img_tuple_identifier,
        ArtifactType.TUPLE,
    )
    assert isinstance(extracted_img_tuple.get(), tuple)
    assert len(extracted_img_tuple.get()) == 2
    assert all(
        isinstance(elem, Image.Image) and elem.getbbox() == image_data.getbbox()
        for elem in extracted_img_tuple.get()
    )

    extracted_img_list = data_integration.file(
        img_list_identifier,
        ArtifactType.LIST,
    )
    assert isinstance(extracted_img_list.get(), list)
    assert len(extracted_img_list.get()) == 3
    assert all(
        isinstance(elem, Image.Image) and elem.getbbox() == image_data.getbbox()
        for elem in extracted_img_list.get()
    )

    # As a trick, we can check that we serialized these objects in the appropriate dictionary format by deserializing them
    # as bytes and manually unpickling in this test case.
    img_tuple_as_bytes = data_integration.file(
        img_tuple_identifier,
        ArtifactType.BYTES,
    )
    import cloudpickle as pickle

    custom_dict = pickle.loads(img_tuple_as_bytes.get())
    assert "aqueduct_serialization_types" in custom_dict
    assert "data" in custom_dict
    assert "is_tuple" in custom_dict


def test_s3_table_formats(flow_manager, data_integration):
    hotel_reviews = data_integration.file(
        "hotel_reviews",
        artifact_type=ArtifactType.TABLE,
        format="parquet",
    )

    # Save the table with each of the other different formats.
    _save_artifact_and_check(flow_manager, data_integration, artifact=hotel_reviews, format="csv")
    _save_artifact_and_check(flow_manager, data_integration, artifact=hotel_reviews, format="json")
    _save_artifact_and_check(
        flow_manager, data_integration, artifact=hotel_reviews, format="parquet"
    )


def test_s3_table_fetch_with_merge(client, data_integration):
    hotel_reviews = data_integration.file(
        "hotel_reviews", artifact_type=ArtifactType.TABLE, format="parquet"
    )
    customers = data_integration.file(
        "customers",
        artifact_type=ArtifactType.TABLE,
        format="parquet",
    )
    expected_merged_data = pd.concat([hotel_reviews.get(), customers.get()], ignore_index=True)

    merged = data_integration.file(
        ["hotel_reviews", "customers"],
        artifact_type=ArtifactType.TABLE,
        format="parquet",
        merge=True,
    )
    assert merged.type() == ArtifactType.TABLE
    assert merged.get().equals(expected_merged_data)


def test_s3_fetch_directory_mixed(flow_manager, data_integration):
    """Create a random directory name and save a table and non-tabular artifact into it, and
    check that a directory fetch will return a tuple of the contents."""
    dir_name = generate_object_name()

    # Order hotel_reviews to be listed before customers by ordering the paths alphabetically.
    hotel_reviews_table_name, customers_table_name = sorted(
        [generate_table_name(), generate_table_name()]
    )

    # Write a tabular artifact into the directory.
    hotel_reviews = data_integration.file(
        "hotel_reviews", artifact_type=ArtifactType.TABLE, format="parquet"
    )
    _save_artifact_and_check(
        flow_manager,
        data_integration,
        artifact=hotel_reviews,
        format="parquet",
        object_identifier="%s/%s" % (dir_name, hotel_reviews_table_name),
    )

    # Check that the artifact can be fetched by directory search.
    dir_contents = data_integration.file(
        dir_name + "/",
        artifact_type=ArtifactType.TABLE,
        format="parquet",
    )
    assert dir_contents.type() == ArtifactType.TUPLE
    assert dir_contents.get()[0].equals(hotel_reviews.get())

    customers = data_integration.file(
        "customers",
        artifact_type=ArtifactType.TABLE,
        format="parquet",
    )
    _save_artifact_and_check(
        flow_manager,
        data_integration,
        artifact=customers,
        format="parquet",
        object_identifier="%s/%s" % (dir_name, customers_table_name),
    )

    # Check that both artifacts can be fetched by directory search.
    dir_contents = data_integration.file(
        dir_name + "/",
        artifact_type=ArtifactType.TABLE,
        format="parquet",
    )

    assert dir_contents.type() == ArtifactType.TUPLE
    dir_contents_data = dir_contents.get()
    assert len(dir_contents_data) == 2
    assert dir_contents_data[0].equals(hotel_reviews.get()) and dir_contents_data[1].equals(
        customers.get()
    )


def test_s3_fetch_directory_with_merge(flow_manager, data_integration):
    dir_name = generate_object_name()

    # Order hotel_reviews to be listed before customers by ordering the paths alphabetically.
    hotel_reviews_table_name, customers_table_name = sorted(
        [generate_table_name(), generate_table_name()]
    )

    # Write two tables into the directory.
    hotel_reviews = data_integration.file(
        "hotel_reviews", artifact_type=ArtifactType.TABLE, format="parquet"
    )
    _save_artifact_and_check(
        flow_manager,
        data_integration,
        artifact=hotel_reviews,
        format="parquet",
        object_identifier="%s/%s" % (dir_name, hotel_reviews_table_name),
    )

    customers = data_integration.file(
        "customers", artifact_type=ArtifactType.TABLE, format="parquet"
    )
    _save_artifact_and_check(
        flow_manager,
        data_integration,
        artifact=customers,
        format="parquet",
        object_identifier="%s/%s" % (dir_name, customers_table_name),
    )

    expected_merged_data = pd.concat([hotel_reviews.get(), customers.get()], ignore_index=True)
    dir_contents_merged = data_integration.file(
        dir_name + "/",
        artifact_type=ArtifactType.TABLE,
        format="parquet",
        merge=True,
    )
    assert dir_contents_merged.type() == ArtifactType.TABLE
    assert expected_merged_data.equals(dir_contents_merged.get())


def test_s3_fetch_directory_with_delete(flow_manager, data_integration):
    """Create a random directory name and save a table artifact into it. Delete the workflow,
    and include that object in the deletion. After that check if the directory is empty."""
    dir_name = generate_object_name()

    # Order hotel_reviews to be listed before customers by ordering the paths alphabetically.
    hotel_reviews_table_name = generate_table_name()

    # Write a tabular artifact into the directory.
    hotel_reviews = data_integration.file(
        "hotel_reviews", artifact_type=ArtifactType.TABLE, format="parquet"
    )
    flow = _save_artifact_and_check(
        flow_manager,
        data_integration,
        artifact=hotel_reviews,
        format="parquet",
        object_identifier="%s/%s" % (dir_name, hotel_reviews_table_name),
    )

    # Check that the artifact can be fetched by directory search.
    dir_contents = data_integration.file(
        dir_name + "/",
        artifact_type=ArtifactType.TABLE,
        format="parquet",
    )
    assert dir_contents.type() == ArtifactType.TUPLE
    assert dir_contents.get()[0].equals(hotel_reviews.get())

    # Delete the workflow and include artifact into deletion.
    tables = flow.list_saved_objects()
    flow_manager._client.delete_flow(
        flow_id=str(flow.id()),
        saved_objects_to_delete=tables,
        force=True,
    )
    # Check if the directory is an empty directory.
    with pytest.raises(
        AqueductError, match="Given path to S3 directory '%s/' does not exist." % dir_name
    ):
        dir_contents = data_integration.file(
            dir_name + "/",
            artifact_type=ArtifactType.TABLE,
            format="parquet",
        )


def test_s3_non_tabular_fetch(client, flow_manager, data_integration):
    string_data = "This is a string."
    non_tabular_artifact = client.create_param("Non-Tabular Data", default=string_data)
    assert non_tabular_artifact.type() == ArtifactType.STRING

    _save_artifact_and_check(
        flow_manager,
        data_integration,
        artifact=non_tabular_artifact,
        format=None,
    )


def test_s3_fetch_multiple_files(client, flow_manager, data_integration):
    hotel_reviews = data_integration.file(
        "hotel_reviews",
        artifact_type=ArtifactType.TABLE,
        format="parquet",
    )
    customers = data_integration.file(
        "customers",
        artifact_type=ArtifactType.TABLE,
        format="parquet",
    )

    multi_table_artifact = data_integration.file(
        ["hotel_reviews", "customers"], artifact_type=ArtifactType.TABLE, format="parquet"
    )
    assert multi_table_artifact.type() == ArtifactType.TUPLE
    multi_table_data = multi_table_artifact.get()
    assert len(multi_table_data) == 2
    assert multi_table_data[0].equals(hotel_reviews.get())
    assert multi_table_data[1].equals(customers.get())

    # Test successful multiple file fetch of non-tabular data.
    non_tabular_data_list_1 = client.create_param("List Param 1", default=[1, 2, 3])
    non_tabular_data_list_2 = client.create_param("List Param 2", default=[4, 5, 6])
    _save_artifact_and_check(
        flow_manager,
        data_integration,
        artifact=non_tabular_data_list_1,
        format=None,
    )
    _save_artifact_and_check(
        flow_manager,
        data_integration,
        artifact=non_tabular_data_list_2,
        format=None,
    )

    multi_data_artifact = data_integration.file(
        [
            artifact_id_to_saved_identifier[str(non_tabular_data_list_1.id())],
            artifact_id_to_saved_identifier[str(non_tabular_data_list_2.id())],
        ],
        artifact_type=ArtifactType.LIST,
        format=None,
    )
    assert multi_data_artifact.type() == ArtifactType.TUPLE
    assert multi_data_artifact.get() == ([1, 2, 3], [4, 5, 6])


def test_s3_fetch_single_file_as_list(data_integration):
    """Check that fetching a single file as a list of paths will return a Tuple artifact."""
    hotel_reviews = data_integration.file(
        ["hotel_reviews"],
        artifact_type=ArtifactType.TABLE,
        format="parquet",
    )
    assert hotel_reviews.type() == ArtifactType.TUPLE
    check_hotel_reviews_table_data(hotel_reviews.get()[0])


def test_s3_artifact_with_custom_metadata(
    flow_manager,
    data_integration,
):
    # TODO: validate custom descriptions once we can fetch descriptions easily.
    artifact = data_integration.file(
        "hotel_reviews",
        artifact_type=ArtifactType.TABLE,
        format="parquet",
        name="Test Artifact",
        description="This is a description",
    )
    assert artifact.name() == "Test Artifact artifact"

    flow = flow_manager.publish_flow_test(artifact)
    check_artifact_was_computed(flow, "Test Artifact artifact")


def test_s3_save_with_overwrite(flow_manager, data_integration):
    """Check that we always replace objects that already exist at the filepath."""
    path = generate_table_name()

    hotel_reviews = data_integration.file(
        "hotel_reviews",
        artifact_type=ArtifactType.TABLE,
        format="parquet",
    )
    assert hotel_reviews.type() == ArtifactType.TABLE
    _save_artifact_and_check(
        flow_manager,
        data_integration,
        artifact=hotel_reviews,
        format="csv",
        object_identifier=path,
    )

    # Customers will overwrite the existing hotel_reviews data.
    customers = data_integration.file(
        "customers",
        artifact_type=ArtifactType.TABLE,
        format="parquet",
    )
    assert customers.type() == ArtifactType.TABLE
    _save_artifact_and_check(
        flow_manager,
        data_integration,
        artifact=customers,
        format="csv",
        object_identifier=path,
    )


def test_s3_basic_fetch_failure(client, data_integration):
    # Fetch a path that does not exist will fail.
    with pytest.raises(AqueductError, match="The specified key does not exist."):
        data_integration.file("asdlkf", artifact_type=ArtifactType.TABLE, format="parquet")

    # Fetch a path to directory that does not exist will fail.
    with pytest.raises(AqueductError, match="Given path to S3 directory 'asdlkf/' does not exist."):
        data_integration.file("asdlkf/", artifact_type="bytes")

    # Fetch an artifact with the wrong artifact type.
    with pytest.raises(
        AqueductError, match="The file at path `.*` is not a valid ArtifactType.DICT object."
    ):
        data_integration.file("hotel_reviews", artifact_type=ArtifactType.DICT)

    # Fetch a table artifact with the wrong format.
    with pytest.raises(
        AqueductError,
        match="The file at path `hotel_reviews` is not a valid ArtifactType.TABLE object. \\(with S3 file format `CSV`\\)",
    ):
        data_integration.file("hotel_reviews", artifact_type=ArtifactType.TABLE, format="csv")

    # Fetch a table artifact with an invalid format.
    with pytest.raises(
        InvalidUserArgumentException, match="Unsupported S3 file format `different format`."
    ):
        data_integration.file(
            "hotel_reviews", artifact_type=ArtifactType.TABLE, format="different format"
        )


def test_s3_multi_fetch_failure(client, flow_manager, data_integration):
    # Save a non-tabular artifact.
    non_tabular_path = generate_object_name()
    string_data = "This is a string."
    non_tabular_artifact = client.create_param("Non-Tabular Data", default=string_data)
    assert non_tabular_artifact.type() == ArtifactType.STRING
    _save_artifact_and_check(
        flow_manager,
        data_integration,
        artifact=non_tabular_artifact,
        format=None,
        object_identifier=non_tabular_path,
    )

    # Fetch multiple files of different underlying types.
    with pytest.raises(
        AqueductError,
        match="The file at path `.*` is not a valid ArtifactType.TABLE object. \\(with S3 file format `Parquet`\\)",
    ):
        data_integration.file(
            ["hotel_reviews", non_tabular_path], artifact_type=ArtifactType.TABLE, format="parquet"
        )

    # Fetch and merge multiple files of different underlying types.
    with pytest.raises(
        AqueductError,
        match="The file at path `.*` is not a valid ArtifactType.TABLE object. \\(with S3 file format `Parquet`\\)",
    ):
        data_integration.file(
            ["hotel_reviews", non_tabular_path],
            artifact_type=ArtifactType.TABLE,
            format="parquet",
            merge=True,
        )

    # Fetch multiple files, but one of the files is a directory name.
    with pytest.raises(
        AqueductError, match="Each key in the list must not be a directory, found dir_name/."
    ):
        data_integration.file(
            ["hotel_reviews", "dir_name/"], artifact_type=ArtifactType.TABLE, format="parquet"
        )


def test_s3_save_failure(client, data_integration):
    # Save a non-tabular artifact with a table format field.
    string_data = "This is a string."
    non_tabular_artifact = client.create_param("Non-Tabular Data", default=string_data)
    with pytest.raises(
        InvalidUserArgumentException,
        match="A `format` argument should only be supplied for saving table artifacts.",
    ):
        save(data_integration, non_tabular_artifact, generate_object_name(), format="json")

    # Save a table artifact without a format field.
    hotel_reviews = data_integration.file(
        "hotel_reviews",
        artifact_type=ArtifactType.TABLE,
        format="parquet",
    )
    with pytest.raises(
        InvalidUserArgumentException,
        match="You must supply a file format when saving tabular data into S3 integration",
    ):
        save(data_integration, hotel_reviews, generate_table_name(), format=None)

    # Save an artifact with a completely wrong format.
    with pytest.raises(
        InvalidUserArgumentException, match="Unsupported S3 file format `wrong format`."
    ):
        save(data_integration, hotel_reviews, generate_table_name(), format="wrong format")
