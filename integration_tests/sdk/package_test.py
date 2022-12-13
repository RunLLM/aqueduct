import aqueduct

# runs a test to ensure all APIs are accessible from `aqueduct` package.
def test_package_methods():
    # schedule related
    aqueduct.DayOfMonth
    aqueduct.DayOfWeek
    aqueduct.Hour
    aqueduct.Minute
    aqueduct.hourly
    aqueduct.daily
    aqueduct.monthly
    aqueduct.weekly

    # decorators
    aqueduct.op
    aqueduct.check
    aqueduct.metric

    # decorators related
    aqueduct.CheckSeverity
    aqueduct.LoadUpdateMode

    # notebook related
    aqueduct.get_apikey
    aqueduct.infer_requirements
    aqueduct.global_config
    aqueduct.to_operator