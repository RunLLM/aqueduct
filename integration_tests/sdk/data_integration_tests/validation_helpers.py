from aqueduct.constants.enums import ArtifactType


def check_hotel_reviews_table_artifact(hotel_reviews_artifact):
    """Use to validate hotel_reviews table data is correct across all data integrations."""
    assert hotel_reviews_artifact.type() == ArtifactType.TABLE
    check_hotel_reviews_table_data(hotel_reviews_artifact.get())


def check_hotel_reviews_table_data(hotel_reviews_data):
    assert list(hotel_reviews_data) == [
        "hotel_name",
        "review_date",
        "reviewer_nationality",
        "review",
    ]
    assert hotel_reviews_data.shape[0] == 100
