import time
from typing import List, Union

import torch
from fastchat.conversation import get_default_conv_template
from fastchat.serve.inference import load_model, compute_skip_echo_len

default_max_gpu_memory = "13GiB"
default_temperature = 0.7
default_max_new_tokens = 1024


class Config:
    def __init__(
        self,
        max_gpu_memory: str,
        temperature: float,
        max_new_tokens: int,
    ):
        self.llama_model_path = "aleksickx/llama-7b-hf"
        self.model_path = "/vicuna-7b"
        self.device = "cuda"
        self.num_gpus = "1"
        self.debug = False
        self.load_8bit = False

        self.max_gpu_memory = default_max_gpu_memory
        self.temperature = temperature
        self.max_new_tokens = max_new_tokens

    def describe(self) -> str:
        print("Running Vicuna 7B with the following config:")
        attrs = {
            "max_gpu_memory": self.max_gpu_memory,
            "temperature": self.temperature,
            "max_new_tokens": self.max_new_tokens,
        }
        print("\n".join([f"{attr}: {value}" for attr, value in attrs.items()]))


def download_llama_7b(llama_model_path: str):
    from huggingface_hub import snapshot_download

    print("Downloading LLaMA 7B...")
    snapshot_download(
        repo_id=llama_model_path,
        local_dir="/llama-7b",
        local_dir_use_symlinks=False,
    )


def convert_weight():
    import subprocess

    cmd = [
        "python3",
        "-m",
        "fastchat.model.apply_delta",
        "--base",
        "/llama-7b",
        "--target",
        "/vicuna-7b",
        "--delta",
        "lmsys/vicuna-7b-delta-v1.1",
    ]

    print("Converting LLaMA weights to Vicuna weights...")
    print(subprocess.check_output(cmd))


@torch.inference_mode()
def generate(
    messages: Union[str, List[str]],
    max_gpu_memory: str = default_max_gpu_memory,
    temperature: float = default_temperature,
    max_new_tokens: int = default_max_new_tokens,
) -> Union[str, List[str]]:
    """Invoke the Vicuna 7B model to generate responses.

    Args:
        messages (Union[str, List[str]]): The message(s) to generate responses for.
        max_gpu_memory (str, optional): The maximum amount of GPU memory to use. Default: "13GiB"
        temperature (float, optional): The temperature to use for sampling. Default: 0.7
        max_new_tokens (int, optional): The maximum number of tokens to generate. Default: 1024

    Examples:
        >>> from aqueduct_llm import vicuna_7b
        >>> vicuna_7b.generate("What's the best LLM?", max_gpu_memory="13GiB", temperature=0.7, max_new_tokens=1024)
        "Vicuna 7B is the best LLM!"
    """
    config = Config(
        max_gpu_memory=max_gpu_memory,
        temperature=temperature,
        max_new_tokens=max_new_tokens,
    )
    config.describe()

    download_llama_7b(config.llama_model_path)
    convert_weight()

    if isinstance(messages, str):
        messages = [messages]
    elif isinstance(messages, list):
        if not all(isinstance(message, str) for message in messages):
            raise Exception("The elements in the list must be of type string.")
    else:
        raise Exception("Input must be a string or a list of strings.")

    if isinstance(messages, str):
        messages = [messages]

    print("Loading model...")
    start_time = time.time()

    model, tokenizer = load_model(
        config.model_path,
        config.device,
        config.num_gpus,
        config.max_gpu_memory,
        config.load_8bit,
        debug=config.debug,
    )

    print("Finished loading model.")
    end_time = time.time()
    time_taken = end_time - start_time

    print(f"Time taken: {time_taken:.5f} seconds")

    results = []
    for message in messages:
        msg = message

        conv = get_default_conv_template(config.model_path).copy()
        conv.append_message(conv.roles[0], msg)
        conv.append_message(conv.roles[1], None)
        prompt = conv.get_prompt()

        inputs = tokenizer([prompt])
        output_ids = model.generate(
            torch.as_tensor(inputs.input_ids).cuda(),
            do_sample=True,
            temperature=config.temperature,
            max_new_tokens=config.max_new_tokens,
        )
        outputs = tokenizer.batch_decode(output_ids, skip_special_tokens=True)[0]
        skip_echo_len = compute_skip_echo_len(config.model_path, conv, prompt)
        outputs = outputs[skip_echo_len:]

        results.append(outputs)

    return results[0] if len(results) == 1 else results
