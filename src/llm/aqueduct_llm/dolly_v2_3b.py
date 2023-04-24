import time
from typing import List, Union

import torch
from aqueduct_llm.utils.dolly_instruct_pipeline import InstructionTextGenerationPipeline
from transformers import AutoModelForCausalLM, AutoTokenizer

default_do_sample = True
default_max_new_tokens = 256
default_top_p = 0.92
default_top_k = 0


class Config:
    def __init__(
        self,
        do_sample: bool,
        max_new_tokens: int,
        top_p: float,
        top_k: int,
    ):
        self.model_path = "databricks/dolly-v2-3b"
        self.do_sample = do_sample
        self.max_new_tokens = max_new_tokens
        self.top_p = top_p
        self.top_k = top_k

    def describe(self) -> str:
        print("Running Dolly V2 3B with the following config:")
        attrs = {
            "do_sample": self.do_sample,
            "max_new_tokens": self.max_new_tokens,
            "top_p": self.top_p,
            "top_k": self.top_k,
        }
        print("\n".join([f"{attr}: {value}" for attr, value in attrs.items()]))


def generate(
    messages: Union[str, List[str]],
    do_sample: bool = default_do_sample,
    max_new_tokens: int = default_max_new_tokens,
    top_p: float = default_top_p,
    top_k: int = default_top_k,
) -> Union[str, List[str]]:
    """Invoke the Dolly V2 3B model to generate responses.

    Args:
        messages (Union[str, List[str]]): The message(s) to generate responses for.
        do_sample (bool, optional): Whether or not to use sampling. Defaults to True. Default: True
        max_new_tokens (int, optional): Max new tokens after the prompt to generate. Default: 256
        top_p (float, optional): If set to float < 1, only the smallest set of most probable tokens with
            probabilities that add up to top_p or higher are kept for generation. Default: 0.92.
        top_k (int, optional): The number of highest probability vocabulary tokens to keep for top-k-filtering.
            Default: 0.

    Examples:
        >>> from aqueduct_llm import dolly_v2_3b
        >>> dolly_v2_3b.generate("What's the best LLM?", do_sample=True, max_new_tokens=256, top_p=0.92, top_k=0)
        "Dolly V2 3B is the best LLM!"
    """
    config = Config(do_sample=do_sample, max_new_tokens=max_new_tokens, top_p=top_p, top_k=top_k)
    config.describe()

    if isinstance(messages, str):
        messages = [messages]
    elif isinstance(messages, list):
        if not all(isinstance(message, str) for message in messages):
            raise Exception("The elements in the list must be of type string.")
    else:
        raise Exception("Input must be a string or a list of strings.")

    if isinstance(messages, str):
        messages = [messages]

    print("Downloading and loading model...")
    start_time = time.time()

    tokenizer = AutoTokenizer.from_pretrained(config.model_path, padding_side="left")
    model = AutoModelForCausalLM.from_pretrained(
        config.model_path, device_map="auto", torch_dtype=torch.bfloat16
    )

    print("Finished loading model.")
    end_time = time.time()
    time_taken = end_time - start_time

    print(f"Time taken: {time_taken:.5f} seconds")

    generate_text = InstructionTextGenerationPipeline(
        model=model,
        tokenizer=tokenizer,
        do_sample=config.do_sample,
        max_new_tokens=config.max_new_tokens,
        top_p=config.top_p,
        top_k=config.top_k,
    )

    results = []
    for message in messages:
        res = generate_text(message)
        results.append(res[0]["generated_text"])

    return results[0] if len(results) == 1 else results
