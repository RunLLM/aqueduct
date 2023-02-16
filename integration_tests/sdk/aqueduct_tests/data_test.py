from PIL import Image

from aqueduct import op
from sdk.shared.flow_helpers import publish_flow_test


def test_expanded_collection_data_type(client, flow_name, engine):
    """Test a list and tuple of images."""

    # Current working directory is one level above.
    image_data = Image.open("aqueduct_tests/data/aqueduct.jpg", "r")

    @op(outputs=["list output"])
    def list_of_images(image: Image):
        return [image, image, image]

    list_output = list_of_images(image_data)
    assert isinstance(list_output.get(), list)
    assert len(list_output.get()) == 3
    assert all(isinstance(elem, Image.Image) for elem in list_output.get())

    @op(outputs=["tuple output"])
    def tuple_of_images(image: Image):
        return (image, image, image)

    tuple_output = tuple_of_images(image_data)
    assert isinstance(tuple_output.get(), tuple)
    assert len(tuple_output.get()) == 3
    assert all(isinstance(elem, Image.Image) for elem in tuple_output.get())

    flow = publish_flow_test(
        client,
        name=flow_name(),
        artifacts=[list_output, tuple_output],
        engine=engine,
    )
    flow_list_output = flow.latest().artifact("list output")
    assert len(flow_list_output.get()) == 3
    assert all(isinstance(elem, Image.Image) for elem in flow_list_output.get())
    flow_tuple_output = flow.latest().artifact("tuple output")
    assert len(flow_tuple_output.get()) == 3
    assert all(isinstance(elem, Image.Image) for elem in flow_tuple_output.get())


def test_expanded_collection_data_type_mixed(client):
    """Check that we can handle pickled lists with a variety of data types."""

    # Current working directory is one level above.
    image_data = Image.open("aqueduct_tests/data/aqueduct.jpg", "r")

    @op
    def foo(image: Image):
        return (b"this is content", image)

    output = foo(image_data)
    assert isinstance(output.get(), tuple)
    assert len(output.get()) == 2
    assert output.get()[0] == b"this is content"
    assert isinstance(output.get()[1], Image.Image)
