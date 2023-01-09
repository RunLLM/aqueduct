from typing import List

from aqueduct.constants.enums import RelationalDBServices, ServiceType

# We limit the number of rows to speed up a database writes a littel bit.
SHORT_SENTIMENT_SQL_QUERY = "select * from hotel_reviews limit 5"


def all_relational_DBs() -> List[ServiceType]:
    return [ServiceType(relational_service.value) for relational_service in RelationalDBServices]
