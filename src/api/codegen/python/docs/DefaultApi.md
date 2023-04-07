# swagger_client.DefaultApi

All URIs are relative to *http://localhost:8080/api/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**workflow_get**](DefaultApi.md#workflow_get) | **GET** /workflow/{workflowID} | get metadata of a workflow

# **workflow_get**
> Workflow workflow_get(workflow_id, api_key)

get metadata of a workflow

### Example
```python
from __future__ import print_function
import time
import swagger_client
from swagger_client.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = swagger_client.DefaultApi()
workflow_id = '38400000-8cf0-11bd-b23e-10b96e4ef00d' # str | the ID of workflow object
api_key = 'api_key_example' # str | the user's API Key

try:
    # get metadata of a workflow
    api_response = api_instance.workflow_get(workflow_id, api_key)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling DefaultApi->workflow_get: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **workflow_id** | [**str**](.md)| the ID of workflow object | 
 **api_key** | **str**| the user&#x27;s API Key | 

### Return type

[**Workflow**](Workflow.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

