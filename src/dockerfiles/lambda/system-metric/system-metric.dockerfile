FROM public.ecr.aws/lambda/python:3.8

# Copy function code
COPY lambda/system-metric/system_metric.py ${LAMBDA_TASK_ROOT}

# Install the function's dependencies using file requirements.txt
# from your project folder.

COPY lambda/requirements.txt  .
RUN  pip3 install -r requirements.txt --target "${LAMBDA_TASK_ROOT}"

RUN pip3 install --index-url https://test.pypi.org/simple/ --extra-index-url https://pypi.org/simple/ aqueduct-ml-test-hari==0.0.1 --target "${LAMBDA_TASK_ROOT}"

# Set the CMD to your handler (could also be done as a parameter override outside of the Dockerfile)
CMD [ "system_metric.handler" ]