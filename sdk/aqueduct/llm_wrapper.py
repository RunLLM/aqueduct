from aqueduct.decorator import op
import aqueduct
from aqueduct.error import InvalidUserArgumentException

def generate_llama_7b(params):
    def use_llama(input):
        from aqueduct_llm import llama_7b
        return llama_7b.generate(input, **params)
    return use_llama

def llm_op(name, params={}, engine=None):
    kwargs = {}
    if engine is not None:
        kwargs['engine'] = engine
        kwargs['resources'] = {
                'memory': '16GB',
                'gpu_resource_name': 'nvidia.com/gpu',
            }
    else:
        raise InvalidUserArgumentException("engine cannot be None")

    if name == "llama_7b":
        return op(
            use=aqueduct.llm,
            requirements=[],
            **kwargs,
        )(generate_llama_7b(params))
    else:
        raise InvalidUserArgumentException(f"Unknown LLM model {name}")