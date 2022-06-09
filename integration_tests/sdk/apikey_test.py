from aqueduct import apikey

def test_apikey_local_server(sp_client):
    sdk_api_key = apikey()
    assert(sdk_api_key == sp_client._api_client.api_key)

