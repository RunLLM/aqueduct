from typing import Optional

from aqueduct.models.operators import LLMSpec

class Vicuna7b(LLMSpec):
    def __init__(
        self, max_gpu_memory: Optional[str] = None,
    ):
        super().__init__(
            name='vicuna_7b',
            requires_gpu=True,
            min_required_memory=16384, # 16GB
            config={},
        )

        if max_gpu_memory is not None:
            self.config['AQUEDUCT_VICUNA_7B_MAX_GPU_MEMORY'] = max_gpu_memory


class DollyV23b(LLMSpec):
    def __init__(
        self,
    ):
        super().__init__(
            name='dolly_v2_3b',
            requires_gpu=True,
            min_required_memory=8192, # 8GB
            config={},
        )


class DollyV27b(LLMSpec):
    def __init__(
        self,
    ):
        super().__init__(
            name='dolly_v2_7b',
            requires_gpu=True,
            min_required_memory=16384, # 16GB
            config={},
        )


class Llama7b(LLMSpec):
    def __init__(
        self,
    ):
        super().__init__(
            name='llama_7b',
            requires_gpu=True,
            min_required_memory=16384, # 16GB
            config={},
        )
