import time
from typing import List, Union

import torch
from transformers import LlamaForCausalLM, LlamaTokenizer

default_max_length = 100


class Config:
    def __init__(self, max_length: int):
        self.model_path = "aleksickx/llama-7b-hf"
        self.device = "cuda"
        self.max_length = max_length

    def describe(self) -> str:
        print("Running LLaMA 7B with the following config:")
        attrs = {
            "max_length": self.max_length,
        }
        print("\n".join([f"{attr}: {value}" for attr, value in attrs.items()]))


def generate(
    messages: Union[str, List[str]],
    max_length: int = default_max_length,
) -> Union[str, List[str]]:
    """Invoke the LLaMA 7B model to generate responses.

    Args:
        messages (Union[str, List[str]]): The message(s) to generate responses for.
        max_length (int, optional): The maximum length of the generated response. Default: 100

    Examples:
        >>> from aqueduct_llm import llama_7b
        >>> llama_7b.generate("What's the best LLM?", max_length=100)
        "LLaMA 7B is the best LLM!"
    """
    config = Config(max_length=max_length)
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

    tokenizer = LlamaTokenizer.from_pretrained(config.model_path)
    model = LlamaForCausalLM.from_pretrained(config.model_path, torch_dtype=torch.bfloat16).to(
        config.device
    )

    print("Finished loading model.")
    end_time = time.time()
    time_taken = end_time - start_time

    print(f"Time taken: {time_taken:.5f} seconds")

    results = []
    for message in messages:
        batch = tokenizer(message, return_tensors="pt", add_special_tokens=False)

        batch = {k: v.to(config.device) for k, v in batch.items()}
        generated = model.generate(batch["input_ids"], max_length=config.max_length)

        results.append(tokenizer.decode(generated[0]))

    return results[0] if len(results) == 1 else results
