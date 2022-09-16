FROM public.ecr.aws/lambda/python:3.7

# Copy function code
COPY lambda/function/function.py ${LAMBDA_TASK_ROOT}

# Install the function's dependencies using file requirements.txt
# from your project folder.

COPY lambda/function/requirements-37.txt  .
RUN  pip3 install -r requirements-37.txt --target "${LAMBDA_TASK_ROOT}"

# Set the CMD to your handler (could also be done as a parameter override outside of the Dockerfile)
CMD [ "function.handler" ]